package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/photos"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type PhotoHandler struct {
	s *server
}

func (ph *PhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	offset := 0
	count := 30
	ascending := true
	if of, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = of
	} else if r.URL.Query().Get("offset") != "" {
		log.WithError(err).Error("error parsing 'offset' param")
	}
	if cnt, err := strconv.Atoi(r.URL.Query().Get("count")); err == nil {
		count = cnt
	} else if r.URL.Query().Get("count") != "" {
		log.WithError(err).Error("error parsing 'count' param")
	}
	if asc := r.URL.Query().Get("sort_dir"); asc == "desc" {
		ascending = false
	}
	ps, dirs, err := ph.s.actions.List(r.Context(), ListReq{
		Offset: offset,
		Count:  count,
		Asc:    ascending,
		Sort:   r.URL.Query().Get("sort"),
		Path:   r.URL.Query().Get("path"),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, r, struct {
		Photos []photos.Photo `json:"photos"`
		Dirs   []string       `json:"dirs"`
	}{ps, dirs})
}

func (ph *PhotoHandler) Get(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	path := chi.URLParam(r, "*")
	photoDir := r.Context().Value("photoDir").(string)
	srcpath := photoDir + "/" + path
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
	photoDir := r.Context().Value("photoDir").(string)
	thumbDir := r.Context().Value("thumbDir").(string)

	size := photos.ParseSize(chi.URLParam(r, "size"))
	path := chi.URLParam(r, "*")

	// look for original file
	search := photoDir + "/" + strings.Replace(path, ".jpg", ".*", -1)
	matches, err := filepath.Glob(search)
	log.WithField("search", search).Debug("searching for original file")
	if err != nil {
		log.WithError(err).Error("error looking for original for thumb")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(matches) == 0 {
		log.Debug("original file not found, returning 404")
		http.NotFound(w, r)
		return
	}
	if len(matches) > 2 {
		log.WithField("matches", matches).Warnf("found %d source matches!", len(matches))
	}
	s := SizeReq{
		File:    strings.TrimPrefix(matches[0], photoDir+"/"),
		Size:    size,
		B64:     false,
		Purpose: r.URL.Query().Get("purpose"),
	}

	if _, err := ph.s.actions.GetSize(r.Context(), s); err != nil {
		log.WithError(err).Error("error getting image at size")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fp := thumbDir + "/" + size.String() + "/" + path
	log.WithField("file", fp).Debug("sending thumb file")
	http.ServeFile(w, r, fp)
}

type SockRequest struct {
	Action string                 `json:"action"`
	ID     string                 `json:"_id"`
	Params map[string]interface{} `json:"params"`
}

type SockResponse struct {
	ID    string      `json:"_id"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func (ph *PhotoHandler) Websocket(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	photoDir := r.Context().Value("photoDir").(string)
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

		resp := SockResponse{ID: req.ID}
		switch req.Action {
		case "list", "List":
			log.WithField("request", req).Trace("parsed as list request action")
			offset := 0
			count := 30
			ascending := true
			sort := ""
			path := ""
			if of, ok := req.Params["offset"]; ok {
				if ofint, ok := of.(float64); ok {
					offset = int(ofint)
				} else {
					log.WithField("offset", of)
					resp.Error = "offset expected to be an integer"
					break
				}
			}
			if c, ok := req.Params["count"]; ok {
				if cnt, ok := c.(float64); ok {
					count = int(cnt)
				} else {
					resp.Error = "count expected to be an integer"
					break
				}
			}
			if dir, ok := req.Params["sort_dir"]; ok {
				if sdir, ok := dir.(string); ok {
					ascending = sdir == "asc"
				} else {
					resp.Error = "sort_dir expected to be a string"
					break
				}
			}
			if srt, ok := req.Params["sort"]; ok {
				if ss, ok := srt.(string); ok {
					sort = ss
				} else {
					resp.Error = "sort expected to be a string"
					break
				}
			}
			if pth, ok := req.Params["path"]; ok {
				if p, ok := pth.(string); ok {
					path = p
				} else {
					resp.Error = "path expected to be a string"
					break
				}
			}
			photos, dirs, err := ph.s.actions.List(r.Context(), ListReq{
				Offset: offset,
				Count:  count,
				Asc:    ascending,
				Sort:   sort,
				Path:   path,
			})
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Data = map[string]interface{}{
					"photos": photos,
					"dirs":   dirs,
				}
			}
			log.WithField("resp", resp).Trace("responding to list request")
		case "size", "Size":
			log.WithField("request", req).Trace("parsed size request action")
			sr := SizeReq{}
			if f, ok := req.Params["file"]; !ok {
				resp.Error = "missing file argument"
				break
			} else {
				if fs, ok := f.(string); !ok {
					resp.Error = "file expected to be a string"
					break
				} else {
					sr.File = fs
				}
			}
			if sz, ok := req.Params["size"]; !ok {
				resp.Error = "missing size argument"
				break
			} else {
				if ss, ok := sz.(string); !ok {
					resp.Error = "size expected to be a string"
					break
				} else {
					sr.Size = photos.ParseSize(ss)
				}
			}
			if b64, ok := req.Params["b64"]; ok {
				switch v := b64.(type) {
				case bool:
					sr.B64 = v
				case int:
					sr.B64 = v == 1
				case string:
					sr.B64 = v == "1" || v == "true"
				default:
					sr.B64 = false
				}
			}
			if purpose, ok := req.Params["purpose"]; ok {
				ps, ok := purpose.(string)
				if !ok {
					resp.Error = "purpose expected to be a string"
					break
				}
				sr.Purpose = ps
			}

			data, err := ph.s.actions.GetSize(r.Context(), sr)
			if err != nil {
				resp.Error = err.Error()
				break
			}
			resp.Data = data
		case "meta", "Meta", "META":
			log.WithField("request", req).Trace("photo info request")
			if _, ok := req.Params["file"]; !ok {
				resp.Error = "missing file argument"
				break
			}
			file, ok := req.Params["file"].(string)
			if !ok {
				resp.Error = "file expected to be a string"
				break
			}
			if p, err := photos.FromSrc(r.Context(), photoDir+"/"+file); err != nil {
				log.WithError(err).Error("failed to get meta info")
				resp.Error = "failed to get meta info"
			} else {
				if m, err := p.Meta(); err != nil {
					resp.Error = "failed to load meta info"
				} else {
					resp.Data = m
				}
			}
		case "":
			resp.Error = "missing action"
		default:
			resp.Error = "unknown action: " + req.Action
		}
		wsjson.Write(r.Context(), c, resp)
	}

	c.Close(websocket.StatusNormalClosure, "k im done with you")
}

// ------------ helpers / internal funcs

func thumbExt(filename string) string {
	r := strings.NewReplacer(
		".ARW", ".jpg",
		".CR2", ".jpg",
		".RAW", ".jpg",
	)
	return r.Replace(filename)
}

func gethost(ctx context.Context) string { return ctx.Value("host").(string) }
