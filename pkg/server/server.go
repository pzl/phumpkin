package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/dgraph-io/badger"
	"github.com/go-chi/chi"
	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/photos"
	"github.com/pzl/phumpkin/pkg/resize"
	"github.com/sirupsen/logrus"
)

type OptFunc func(s *server)

type server struct {
	*mstk.Server
	thumbDir     string
	photoDir     string
	dataDir      string
	db           *badger.DB
	assets       http.Handler
	router       *chi.Mux
	PhotoHandler PhotoHandler
	resizer      *resize.Resizer
	actions      Action
	mgr          *photos.Mgr
}

func New(options ...OptFunc) *server {
	s := &server{
		Server:  mstk.NewServer(),
		router:  chi.NewRouter(),
		resizer: resize.New(),
		mgr:     photos.New(),
	}
	s.PhotoHandler.s = s
	s.actions.s = s
	s.Server.Http.Handler = s.router
	for _, o := range options {
		if o != nil {
			o(s)
		}
	}
	return s
}

func (s *server) Start(ctx context.Context) (err error) {

	c := context.WithValue(ctx, "log", s.Log)
	c = context.WithValue(c, "photoDir", s.photoDir)
	c = context.WithValue(c, "dataDir", s.dataDir)
	c = context.WithValue(c, "thumbDir", s.thumbDir)

	db, err := badger.Open(badger.DefaultOptions(s.dataDir))
	if err != nil {
		return err
	}
	s.db = db
	c = context.WithValue(c, "badger", db)

	// set server db before setting up routes, where ctx middleware will pick it up
	s.routes()

	if err := s.mgr.Start(c); err != nil {
		return err
	}
	s.resizer.Start(c)
	return s.Server.Start(c)
}

func (s *server) Shutdown(ctx context.Context) {
	s.db.Close()
	s.Server.Shutdown(ctx)
}

func Addr(addr string) OptFunc      { return func(s *server) { mstk.Addr(addr)(s.Server) } }
func Log(l *logrus.Logger) OptFunc  { return func(s *server) { s.Log = l } }
func Photos(d string) OptFunc       { return func(s *server) { s.photoDir = filepath.Clean(d) } }
func Thumbs(d string) OptFunc       { return func(s *server) { s.thumbDir = filepath.Clean(d) } }
func DataDir(d string) OptFunc      { return func(s *server) { s.dataDir = filepath.Clean(d) } }
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
