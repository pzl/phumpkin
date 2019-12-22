package server

import (
	"context"
	"os"
	"strings"

	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

// @todo: duplicates by XMP
// primary may be <IMG>.ARW.xmp and dupe may be <IMG>_nn.ARW.XMP
func actionList(ctx context.Context, log logrus.FieldLogger, dir string, host string) ([]FileInfo, error) {
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

	log.WithField("photoDir", dir).Debug("scanning photoDir")
	err := walker.WalkWithContext(ctx, dir, func(name string, fi os.FileInfo) error {
		log.WithField("filename", name).Trace("looping over file")

		if name == dir {
			return nil
		}
		if fi.IsDir() { // recurse into dirs, but ignore folder entries themselves
			return nil //filepath.SkipDir
		}
		if strings.HasSuffix(name, ".xmp") {
			return nil
		}
		path := strings.TrimPrefix(name, dir+"/")

		thumbs := make(map[string]Resource)

		for _, s := range Sizes {
			thumbs[s.Name] = Resource{
				URL:    "http://" + host + "/api/v1/thumb/" + s.Name + "/" + thumbExt(path),
				Width:  s.Max, //@todo get from actual
				Height: s.Max, // ^^
			}
		}

		var x darktable.Meta

		if m, err := darktable.ReadXMP(name + ".xmp"); err == nil {
			x = m
		} else {
			log.WithError(err).Error("error reading XMP")
		}
		if _, err := darktable.ReadExif(name); err != nil {
			log.WithError(err).Error("error reading EXIF")
		}

		found <- FileInfo{
			Name:     path,
			Dir:      fi.IsDir(),
			Size:     fi.Size(),
			Rating:   3,
			Tags:     []string{},
			Location: nil,
			XMP:      &x,
			Original: Resource{
				URL:    "http://" + host + "/api/v1/photos/" + path,
				Width:  2000,
				Height: 2000,
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
