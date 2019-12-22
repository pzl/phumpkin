package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

	photos, err := actionList(r.Context(), log, ph.photoDir, r.Host)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, r, struct {
		Photos []FileInfo `json:"photos"`
	}{photos})
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

type SockRequest struct {
	Action string   `json:"action"`
	ID     string   `json:"_id"`
	Params []string `json:"params"`
}

type SockResponse struct {
	ID    string      `json:"_id"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func (ph *PhotoHandler) Websocket(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // @todo: turn off after dev done
	})
	if err != nil {
		log.WithError(err).Error("error when establishing websocket connection")
		return
	}
	defer c.Close(websocket.StatusInternalError, "died.")
	log.Debug("got websocket connection")

	var req SockRequest
	for {
		_, rd, err := c.Reader(r.Context())
		if err != nil {
			// when going to a new page in browser: StatusGoingAway
			// when just close() -- StatusNoStatusRcvd
			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure,
				websocket.StatusGoingAway,
				websocket.StatusNoStatusRcvd:
				log.WithError(err).Info("socket closed")
			default:
				log.WithError(err).Error("socket closed, received unacceptable error")
			}
			return
		}
		if err := json.NewDecoder(rd).Decode(&req); err != nil {
			log.WithError(err).Error("error unmarshalling json")
			wsjson.Write(r.Context(), c, map[string]string{
				"error": "invalid message",
			})
			continue
		}
		log.WithField("request", req).Trace("got websocket message")

		if req.ID == "" {
			wsjson.Write(r.Context(), c, map[string]interface{}{
				"error": "missing ID",
			})
			continue
		}

		switch req.Action {
		case "list", "List":
			log.WithField("request", req).Trace("parsed as list request action")
			photos, err := actionList(r.Context(), log, ph.photoDir, r.Host)
			resp := SockResponse{ID: req.ID}
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Data = photos
			}
			log.WithField("resp", resp).Trace("responding to list request")
			wsjson.Write(r.Context(), c, resp)
		case "":
			wsjson.Write(r.Context(), c, map[string]interface{}{
				"id":    req.ID,
				"error": "missing action",
			})
		default:
			wsjson.Write(r.Context(), c, map[string]interface{}{
				"id":    req.ID,
				"error": "unknown action: " + req.Action,
			})
		}
	}

	c.Close(websocket.StatusNormalClosure, "k im done with you")

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
