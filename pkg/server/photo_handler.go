package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

type PhotoHandler struct {
	photoDir  string
	thumbDir  string
	dataDir   string
	darktable *darktable.Exporter
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
	Rating   int                 `json:"rating"`
	Tags     []string            `json:"tags"`
	XMP      *darktable.Meta     `json:"xmp"`
	Location *Location           `json:"loc"`
	Thumbs   map[string]Resource `json:"thumbs"`
	Original Resource            `json:"original"`
}

func (ph *PhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	/*
		f, err := os.Open(ph.photoDir)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}

		files, err := f.Readdir(0)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
	*/

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

	log.WithField("photoDir", ph.photoDir).Debug("scanning photoDir")
	err := walker.WalkWithContext(r.Context(), ph.photoDir, func(name string, fi os.FileInfo) error {
		log.WithField("filename", name).Trace("looping over file")

		if name == ph.photoDir {
			return nil
		}
		if fi.IsDir() { // recurse into dirs, but ignore folder entries themselves
			return nil //filepath.SkipDir
		}
		if strings.HasSuffix(name, ".xmp") {
			return nil
		}
		path := strings.TrimPrefix(name, ph.photoDir+"/")

		thumbs := make(map[string]Resource)

		for _, s := range Sizes {
			thumbs[s.Name] = Resource{
				URL:    "http://" + r.Host + "/api/v1/thumb/" + s.Name + "/" + thumbExt(path),
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

		found <- FileInfo{
			Name:     path,
			Dir:      fi.IsDir(),
			Size:     fi.Size(),
			Rating:   3,
			Tags:     []string{},
			Location: nil,
			XMP:      &x,
			Original: Resource{
				URL:    "http://" + r.Host + "/api/v1/photos/" + path,
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
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	<-done
	writeJSON(w, r, struct {
		Photos []FileInfo `json:"photos"`
	}{files})

}

func (ph *PhotoHandler) Get(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	path := chi.URLParam(r, "*")
	srcpath := ph.photoDir + "/" + path
	l := log.WithField("file", srcpath)
	l.Debug("source file requested")

	if _, err := os.Stat(srcpath); err == nil {
		_, filename := filepath.Split(path)
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	}
	http.ServeFile(w, r, srcpath)
}
func (ph *PhotoHandler) GetThumb(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)

	size := chi.URLParam(r, "size")
	path := chi.URLParam(r, "*")

	thumbpath := ph.thumbDir + "/" + size + "/" + path

	log.WithFields(logrus.Fields{
		"size":      size,
		"path":      path,
		"thumbpath": thumbpath,
	}).Debug("thumb request")
	l := log.WithField("thumb", thumbpath)

	// look for original file
	search := ph.photoDir + "/" + strings.Replace(path, ".jpg", ".*", -1)
	matches, err := filepath.Glob(search)
	l.WithField("search", search).Debug("searching for original file")
	if err != nil {
		l.WithError(err).Error("error looking for original for thumb")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(matches) == 0 {
		l.Debug("original file not found, returning 404")
		http.NotFound(w, r)
		return
	}
	if len(matches) > 2 {
		l.WithField("matches", matches).Warnf("found %d source matches!", len(matches))
	}

	// check modification times of source image and XMPs
	var src string
	var xmp string
	lastMod := time.Unix(0, 0)
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fi.ModTime().After(lastMod) {
			lastMod = fi.ModTime()
		}

		// grab for later use
		if strings.HasSuffix(strings.ToLower(m), ".xmp") {
			xmp = m
		} else {
			src = m
		}
	}
	l.WithField("mod", lastMod).Trace("last modification time of original source")

	// if thumb doesn't already exist (or original has changed), generate on the fly
	fi, err := os.Stat(thumbpath)
	if os.IsNotExist(err) || lastMod.After(fi.ModTime()) {
		l.Debug("generating thumb on the fly")

		job, err := ph.darktable.Immediate(src, thumbpath, Px(size), darktable.SetXMP(xmp))
		if err != nil {
			l.WithError(err).Error("error starting job")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		select {
		case <-job.Done:
			l.Trace("thumb generation job complete")
		case <-r.Context().Done():
			l.Trace("HTTP client disconnected, stopping immediate thumb request")
			job.Cancel()
			return
		}
	} else if err != nil {
		l.WithError(err).Error("error looking up thumb file")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	l.Debug("sending thumb file")
	http.ServeFile(w, r, thumbpath)
}

// ------------ helpers / internal funcs

type Size struct {
	Name string
	Max  int
}

var Sizes = []Size{
	{"x-small", 10},
	{"small", 200},
	{"medium", 800},
	{"large", 1200},
	{"x-large", 2000},
	{"full", 0},
}

func Px(s string) int {
	for _, n := range Sizes {
		if s == n.Name {
			return n.Max
		}
	}
	return 800
}

func thumbExt(filename string) string {
	r := strings.NewReplacer(
		".ARW", ".jpg",
		".CR2", ".jpg",
		".RAW", ".jpg",
	)
	return r.Replace(filename)
}
