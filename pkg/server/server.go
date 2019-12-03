package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk"
	"github.com/sirupsen/logrus"
)

type OptFunc func(s *server)

type server struct {
	*mstk.Server
	photoDir string
	thumbDir string
	assets   http.Handler
	router   *chi.Mux
}

func New(options ...OptFunc) *server {
	s := &server{
		Server: mstk.NewServer(),
		router: chi.NewRouter(),
	}
	s.Server.Http.Handler = s.router
	for _, o := range options {
		if o != nil {
			o(s)
		}
	}
	return s
}

func (s *server) Start(ctx context.Context) (err error) {
	s.routes()
	return s.Server.Start(ctx)
}

func Addr(addr string) OptFunc      { return func(s *server) { mstk.Addr(addr)(s.Server) } }
func Log(l *logrus.Logger) OptFunc  { return func(s *server) { s.Log = l } }
func Photos(d string) OptFunc       { return func(s *server) { s.photoDir = d } }
func Thumbs(d string) OptFunc       { return func(s *server) { s.thumbDir = d } }
func Assets(h http.Handler) OptFunc { return func(s *server) { s.assets = h } }
