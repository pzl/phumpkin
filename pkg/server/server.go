package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/darktable"
	"github.com/sirupsen/logrus"
)

type OptFunc func(s *server)

type server struct {
	*mstk.Server
	assets       http.Handler
	router       *chi.Mux
	PhotoHandler PhotoHandler
	darktable    *darktable.Exporter
}

func New(options ...OptFunc) *server {
	d := darktable.New()
	s := &server{
		Server: mstk.NewServer(),
		router: chi.NewRouter(),
		PhotoHandler: PhotoHandler{
			darktable: d,
		},
		darktable: d,
	}
	s.Server.Http.Handler = s.router
	for _, o := range options {
		if o != nil {
			o(s)
		}
	}

	d.Log = s.Log
	return s
}

func (s *server) Start(ctx context.Context) (err error) {
	s.routes()
	s.darktable.Start(ctx)
	return s.Server.Start(ctx)
}

func Addr(addr string) OptFunc      { return func(s *server) { mstk.Addr(addr)(s.Server) } }
func Log(l *logrus.Logger) OptFunc  { return func(s *server) { s.Log = l } }
func Photos(d string) OptFunc       { return func(s *server) { s.PhotoHandler.photoDir = filepath.Clean(d) } }
func Thumbs(d string) OptFunc       { return func(s *server) { s.PhotoHandler.thumbDir = filepath.Clean(d) } }
func DataDir(d string) OptFunc      { return func(s *server) { s.PhotoHandler.dataDir = filepath.Clean(d) } }
func Assets(h http.Handler) OptFunc { return func(s *server) { s.assets = h } }

// easy http handler escape
func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(buf.Bytes())
	if err != nil {
		log := logger.GetLog(r)
		log.WithError(err).Error("unable to print JSON")
	}
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeFail(w, code, err.Error())
}

func writeFail(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"error":"` + message + `"}`)); err != nil {
		// get logger?
	}
}
