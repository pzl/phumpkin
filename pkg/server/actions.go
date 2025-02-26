package server

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/photos"
	"github.com/pzl/phumpkin/pkg/resize"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

type Action struct {
	s *server
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
func (a Action) List(ctx context.Context, lr ListReq) ([]photos.Photo, []string, error) {
	log := logger.LogFromCtx(ctx)
	photoDir := ctx.Value("photoDir").(string)

	// receiving channels
	rcvPhoto := make(chan photos.Photo)
	rcvDir := make(chan string)
	done := make(chan struct{}) // collector sync channel

	// collectors for sorting and responding with
	files := make([]photos.Photo, 0, 300)
	dirs := make([]string, 0, 20)

	go func() {
		for d := range rcvDir {
			dirs = append(dirs, d)
		}
		done <- struct{}{}
	}()

	go func() {
		for p := range rcvPhoto {
			files = append(files, p)
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
			rcvDir <- strings.TrimPrefix(name, searchPath+"/")
			return filepath.SkipDir // non-recursive for now
		}

		photo, err := photos.FromSrc(ctx, name)
		if err != nil {
			return err
		}
		rcvPhoto <- photo

		return nil
	}, walker.WithErrorCallback(func(name string, err error) error {
		log.WithError(err).Error("encountered err when walking files")
		return nil
	}))
	// done sending, exit these loops, they will send to done when finished appending
	close(rcvDir)
	close(rcvPhoto)

	if err != nil {
		return nil, nil, err
	}

	<-done // wait for dirs loop
	<-done // wait for photos loop

	ps := PhotoSort(lr.Sort, lr.Asc, lr.Count, lr.Offset, files)
	sort.Strings(dirs)
	return ps, dirs, nil
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
	p, err := photos.FromSrc(ctx, filepath)
	if err != nil {
		return "", err
	}

	var xmp string
	lastMod := p.LastMod()
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
			if err := resize.Quick(src, thumbpath, sr.Size.Int()); err != nil {
				l.WithField("src", src).WithField("dest", thumbpath).WithError(err).Error("error resizing with vips")
				return "", err
			}

		} else {
			// small-or-above request, resize using darktable

			// if using raw file, use XMP as a parameter
			if _, err := os.Stat(src + ".xmp"); err == nil {
				xmp = src + ".xmp"
			}

			opts := make([]resize.JobOpt, 0, 1)
			if src == filepath {
				opts = append(opts, resize.SetXMP(xmp))
			}
			if sr.Size == photos.SizeXL || sr.Size == photos.SizeFull {
				opts = append(opts, resize.SetHQ(true)) // turn on HQ for high px count
			}

			job := a.s.resizer.CreateJob(src, thumbpath, sr.Size.Int(), opts...)
			priority := resize.PR_NORMAL
			switch sr.Purpose {
			case "lazysrc":
				priority = resize.PR_HIGH
			case "viewer":
				priority = resize.PR_IMMEDIATE
			}
			a.s.resizer.Add(job, priority)
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

func PhotoSort(sortby string, asc bool, count int, offset int, ps []photos.Photo) []photos.Photo {
	sort.SliceStable(ps, func(i, j int) bool {
		// @todo: give frontend more control in here with an embedded execution string. JS or lua, etc
		switch strings.ToLower(sortby) {
		case "name", "":
			if asc {
				return ps[i].Relpath() < ps[j].Relpath()
			} else {
				return ps[i].Relpath() > ps[j].Relpath()
			}
		case "date taken":
			srt := false
			ps[i].Ex_if_string("DateTimeOriginal", func(a string) {
				ps[j].Ex_if_string("DateTimeOriginal", func(b string) {
					if asc {
						srt = a < b
					} else {
						srt = a > b
					}
				})
			})
			return srt
		case "rating":
			if asc {
				return ps[i].MetaInt("Rating") < ps[j].MetaInt("Rating")
			} else {
				return ps[i].MetaInt("Rating") > ps[j].MetaInt("Rating")
			}
		}

		return false
	})

	if count <= 0 {
		count = 30
	}
	if offset < 0 {
		offset = 0
	}

	return ps[offset:min(offset+count, len(ps))]
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
