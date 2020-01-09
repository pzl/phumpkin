package server

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/pzl/phumpkin/pkg/photos"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

type Action struct {
	s *server
}

type Resource struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type FileInfo struct {
	Name     string              `json:"name"`
	Size     int64               `json:"size"`
	Dir      bool                `json:"-"`
	Rotation int                 `json:"rotation"`
	Meta     *photos.Meta        `json:"meta"`
	Thumbs   map[photos.Size]Resource `json:"thumbs"`
	Original Resource            `json:"original"`
}

type ListReq struct {
	Offset int
	Count  int
	Sort   string
	Asc    bool
	Path   string
}

// @todo: duplicates by XMP
// primary may be <IMG>.ARW.xmp and dupe may be <IMG>_nn.ARW.XMP
func (a Action) List(ctx context.Context, lr ListReq) ([]FileInfo, []string, error) {
	files := make([]FileInfo, 0, 300)
	log := logger.LogFromCtx(ctx)
	photoDir := ctx.Value("photoDir").(string)

	dirs := make([]string, 0, 20)
	found := make(chan FileInfo)
	done := make(chan struct{})

	go func() {
		for f := range found {
			if f.Dir {
				dirs = append(dirs, f.Name)
				r.Log.WithField("name", f.Name).Trace("received dir")
				continue
			}
			r.Log.WithField("name", f.Name).Trace("received file")
			files = append(files, f)
		}
		done <- struct{}{}
	}()

	searchPath := photoDir
	if lr.Path != "" {
		searchPath = filepath.Join(searchPath, path.Clean(lr.Path))
	}
	log.WithField("searchPath", searchPath).Debug("scanning for photos")
	err := walker.WalkWithContext(ctx, searchPath, func(name string, fi os.FileInfo) error {
		log.WithField("filename", name).Trace("looping over file")

		if name == searchPath {
			return nil
		}
		if strings.HasSuffix(name, ".xmp") {
			return nil
		}
		if fi.IsDir() {
			found <- FileInfo{
				Name: strings.TrimPrefix(name, searchPath+"/"),
				Dir:  true,
			}
			return filepath.SkipDir // non-recursive for now
		}

		relpath := strings.TrimPrefix(name, photoDir+"/")
		meta, err := a.s.mgr.Load(log, relpath)
		if err != nil {
			return err
		}

		// @todo: this doesn't account for darktable crops
		fw, ok := meta.EXIF["ImageWidth"].(float64)
		if !ok {
			fw = 8000
		}
		fh, ok := meta.EXIF["ImageHeight"].(float64)
		if !ok {
			fh = 5320
		}
		w := int(fw)
		h := int(fh)

		// switch width and height if rotated
		rotStrings := map[string]int{
			"Horizontal (normal)":                 0,
			"Mirror vertical":                     1,
			"Mirror horizontal":                   2,
			"Rotate 180":                          3,
			"Mirror horizontal and rotate 270 CW": 4,
			"Rotate 90 CW":                        5,
			"Rotate 270 CW":                       6,
			"Mirror horizontal and rotate 90 CW":  7,
		}

		rotation := 0
		if v, ok := meta.EXIF["Orientation"]; ok {
			if s, ok := v.(string); ok {
				if rot, ok := rotStrings[s]; ok {
					rotation = rot
					if rot > 3 {
						w, h = h, w
					}
				}
			}
		}

		thumbs := make(map[photos.Size]Resource)
		for _, s := range []photos.Size{photos.SizeXS, photos.SizeSmall, photos.SizeMedium, photos.SizeLarge, photos.SizeXL, photos.SizeFull} {
			var rw int
			var rh int

			if w > h {
				rw = int(s)
				rh = int(float64(h) / (float64(w) / float64(rw)))
			} else {
				rh = int(s)
				rw = int(float64(w) / (float64(h) / float64(rh)))
			}

			if s == photos.SizeFull {
				rw = w
				rh = h
			}

			thumbs[s] = Resource{
				URL:    "http://" + r.Host + "/api/v1/thumb/" + s.String() + "/" + thumbExt(relpath),
				Width:  rw,
				Height: rh,
			}
		}

		found <- FileInfo{
			Name:     relpath,
			Dir:      false,
			Size:     fi.Size(),
			Rotation: rotation,
			Meta:     &meta,
			Original: Resource{
				URL:    "http://" + r.Host + "/api/v1/photos/" + relpath,
				Width:  w,
				Height: h,
			},
			Thumbs: thumbs,
		}

		return nil
	}, walker.WithErrorCallback(func(name string, err error) error {
		log.WithError(err).Error("encountered err when walking files")
		return nil
	}))
	close(found)

	if err != nil {
		return nil, nil, err
	}

	<-done

	if lr.Sort != "" {
		sort.SliceStable(files, func(i, j int) bool {
			// @todo: give frontend more control in here with an embedded execution string. JS or lua, etc
			switch strings.ToLower(lr.Sort) {
			case "name":
				if lr.Asc {
					return files[i].Name < files[j].Name
				} else {
					return files[i].Name > files[j].Name
				}
			case "date taken":
				if a, aok := files[i].Meta.EXIF["DateTimeOriginal"]; aok {
					if b, bok := files[j].Meta.EXIF["DateTimeOriginal"]; bok {
						if as, ok := a.(string); ok {
							if bs, ok := b.(string); ok {
								if lr.Asc {
									return as < bs
								} else {
									return as > bs
								}
							}
						}
					}
				}
			case "rating":
				if lr.Asc {
					return files[i].Meta.Rating < files[j].Meta.Rating
				} else {
					return files[i].Meta.Rating > files[j].Meta.Rating
				}
			}

			return false
		})
	}

	if lr.Count <= 0 {
		lr.Count = 30
	}
	if lr.Offset < 0 {
		lr.Offset = 0
	}

	return files[lr.Offset:min(lr.Offset+lr.Count, len(files))], dirs, nil
}

type SizeReq struct {
	File    string
	Size    photos.Size
	B64     bool
	Purpose string
}

func (a Action) GetSize(ctx context.Context, sr SizeReq) (string, error) {
	log := logger.LogFromCtx(ctx)
	photoDir := ctx.Value("photoDir").(string)
	thumbDir := ctx.Value("thumbDir").(string)

	filepath := photoDir + "/" + sr.File
	thumbpath := thumbDir + "/" + sr.Size.String() + "/" + thumbExt(sr.File)

	log.WithFields(logrus.Fields{
		"size": sr.Size,
		"file": sr.File,
	}).Debug("size request")
	l := log.WithField("file", sr.File)

	// check modification times of source image and XMPs
	var xmp string
	lastMod := time.Unix(0, 0)
	srcinfo, err := os.Stat(filepath)
	if err != nil {
		return "", nil
	}
	if srcinfo.ModTime().After(lastMod) {
		lastMod = srcinfo.ModTime()
	}
	if fi, err := os.Stat(filepath + ".xmp"); err == nil {
		xmp = filepath + ".xmp"
		if fi.ModTime().After(lastMod) {
			lastMod = fi.ModTime()
		}
	}
	l.WithField("mod", lastMod).Trace("last modification time of original source")

	// if thumb doesn't already exist (or original has changed), generate on the fly
	fi, err := os.Stat(thumbpath)
	if os.IsNotExist(err) || lastMod.After(fi.ModTime()) {
		l.Debug("generating thumb on the fly")
		src := filepath

		// if there is a larger size thumbnail that is still up-to-date, generate from that.
		// it's quicker than using a huge ARW
		if sr.Size < photos.SizeFull { // "full" will not have anything larger
			for _, s := range []photos.Size{photos.SizeXS, photos.SizeSmall, photos.SizeMedium, photos.SizeLarge, photos.SizeXL, photos.SizeFull} {
				if sr.Size >= s { // skip anything smaller than request
					continue
				}
				bigthumb := thumbDir + "/" + s.String() + "/" + thumbExt(sr.File)
				if ti, err := os.Stat(bigthumb); err == nil {
					if ti.ModTime().After(lastMod) {
						src = bigthumb
						break
					}
				}
			}
		}

		if sr.Size == photos.SizeXS || src != filepath {
			// quick trickery using vips

			l.Trace("resizing with vips")
			if err := photos.Resize(src, thumbpath, sr.Size.Int()); err != nil {
				l.WithField("src", src).WithField("dest", thumbpath).WithError(err).Error("error resizing with vips")
				return "", err
			}

		} else {
			// small-or-above request, resize using darktable

			// if using raw file, use XMP as a parameter
			opts := make([]darktable.JobOpt, 0, 1)
			if src == filepath {
				opts = append(opts, darktable.SetXMP(xmp))
			}

			job := a.s.darktable.CreateJob(src, thumbpath, sr.Size.Int(), opts...)
			priority := darktable.PR_NORMAL
			switch sr.Purpose {
			case "lazysrc":
				priority = darktable.PR_HIGH
			case "viewer":
				priority = darktable.PR_IMMEDIATE
			}
			a.s.darktable.Add(job, priority)
			select {
			case <-job.Done:
				l.Trace("thumb generation job complete")
			case <-ctx.Done():
				l.Trace("HTTP client disconnected, stopping immediate thumb request")
				job.Cancel()
				return "", errors.New("canceled")
			}
		}

	} else if err != nil {
		l.WithError(err).Error("error looking up thumb file")
		return "", err
	}
	l.Debug("sending thumb file")

	if !sr.B64 {
		return "http://" + gethost(ctx) + "/api/v1/thumb/" + sr.Size.String() + "/" + thumbExt(sr.File), nil
	}

	// read into b64
	data, err := ioutil.ReadFile(thumbpath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
