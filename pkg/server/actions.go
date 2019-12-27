package server

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/pzl/phumpkin/pkg/photos"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

type Action struct {
	s *server
}

type Location struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}
type Resource struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type FileInfo struct {
	Name     string              `json:"name"`
	Dir      bool                `json:"dir"`
	Size     int64               `json:"size"`
	Rotation int                 `json:"rotation"`
	Meta     *photos.Meta        `json:"meta"`
	Location *Location           `json:"loc"`
	Thumbs   map[string]Resource `json:"thumbs"`
	Original Resource            `json:"original"`
}

// @todo: duplicates by XMP
// primary may be <IMG>.ARW.xmp and dupe may be <IMG>_nn.ARW.XMP
func (a Action) List(ctx context.Context, log logrus.FieldLogger, host string) ([]FileInfo, error) {
	files := make([]FileInfo, 0, 300)
	found := make(chan FileInfo)
	done := make(chan struct{})

	go func() {
		for f := range found {
			log.WithField("name", f).Trace("received file")
			files = append(files, f)
		}
		done <- struct{}{}
	}()

	log.WithField("photoDir", a.s.photoDir).Debug("scanning photoDir")
	err := walker.WalkWithContext(ctx, a.s.photoDir, func(name string, fi os.FileInfo) error {
		log.WithField("filename", name).Trace("looping over file")

		if name == a.s.photoDir {
			return nil
		}
		if fi.IsDir() { // recurse into dirs, but ignore folder entries themselves
			return nil //filepath.SkipDir
		}
		if strings.HasSuffix(name, ".xmp") {
			return nil
		}
		path := strings.TrimPrefix(name, a.s.photoDir+"/")

		meta, err := a.s.mgr.Load(log, name)
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

		thumbs := make(map[string]Resource)
		for _, s := range Sizes {
			var rw int
			var rh int

			if w > h {
				rw = s.Max
				rh = int(float64(h) / (float64(w) / float64(rw)))
			} else {
				rh = s.Max
				rw = int(float64(w) / (float64(h) / float64(rh)))
			}

			if s.Name == "full" {
				rw = w
				rh = h
			}

			thumbs[s.Name] = Resource{
				URL:    "http://" + host + "/api/v1/thumb/" + s.Name + "/" + thumbExt(path),
				Width:  rw,
				Height: rh,
			}
		}

		found <- FileInfo{
			Name:     path,
			Dir:      fi.IsDir(),
			Size:     fi.Size(),
			Location: nil,
			Rotation: rotation,
			Meta:     &meta,
			Original: Resource{
				URL:    "http://" + host + "/api/v1/photos/" + path,
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
		return nil, err
	}

	<-done
	return files, nil
}

func (a Action) GetSize(ctx context.Context, log logrus.FieldLogger, file string, size string, b64 bool, host string) (string, error) {
	filepath := a.s.photoDir + "/" + file
	thumbpath := a.s.thumbDir + "/" + size + "/" + thumbExt(file)

	log.WithFields(logrus.Fields{
		"size": size,
		"file": file,
	}).Debug("size request")
	l := log.WithField("file", file)

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

		// if there is a larger thumb that is still up-to-date, generate from that.
		// it's quicker than using a huge ARW

		for _, s := range Sizes {
			if size == "full" { // no larger thumb
				continue
			}
			if s.Name == size { // skip yourself
				continue
			}
			if s.Max < Px(size) && s.Max != 0 {
				continue // thumb is smaller, except 'full' size
			}
			bigthumb := a.s.thumbDir + "/" + s.Name + "/" + thumbExt(file)
			if ti, err := os.Stat(bigthumb); err == nil {
				if ti.ModTime().After(srcinfo.ModTime()) {
					src = bigthumb
					break
				}
			}
		}

		// if using raw file, use XMP as a parameter
		opts := make([]darktable.JobOpt, 0, 1)
		if src == filepath {
			opts = append(opts, darktable.SetXMP(xmp))
		}

		job, err := a.s.darktable.Immediate(src, thumbpath, Px(size), opts...)
		if err != nil {
			l.WithError(err).Error("error starting job")
			return "", err
		}
		select {
		case <-job.Done:
			l.Trace("thumb generation job complete")
		case <-ctx.Done():
			l.Trace("HTTP client disconnected, stopping immediate thumb request")
			job.Cancel()
			return "", errors.New("canceled")
		}
	} else if err != nil {
		l.WithError(err).Error("error looking up thumb file")
		return "", err
	}
	l.Debug("sending thumb file")

	if !b64 {
		return "http://" + host + "/api/v1/thumb/" + size + "/" + thumbExt(file), nil
	}

	// read into b64
	data, err := ioutil.ReadFile(thumbpath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
