package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
)

func (s *server) routes() {
	s.router.Use(middleware.RealIP) // X-Forwarded-For
	s.router.Use(middleware.RequestID)
	s.router.Use(InjectHost)
	s.router.Use(s.InjectPaths)
	s.router.Use(middleware.RequestLogger(logger.NewChi(s.Log)))
	s.router.Use(middleware.Heartbeat("/ping"))
	s.router.Use(middleware.Recoverer)

	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, User")
			next.ServeHTTP(w, r)
		})
	})

	s.router.Route("/api/v1", func(v1 chi.Router) {
		v1.Use(mstk.APIVer(1))
		v1.Mount("/photos", s.Photos())
		v1.Mount("/query", s.Queries())
		v1.Mount("/complete/", s.Typeahead())
		v1.Get("/thumb/{size}/*", s.PhotoHandler.GetThumb)
		v1.Get("/ws", s.PhotoHandler.Websocket)

	})

	s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		s.assets.ServeHTTP(w, r)
	})
}

// inject request.Host into context
func InjectHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "host", r.Host)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// inject config paths into context
func (s server) InjectPaths(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "photoDir", s.photoDir)
		ctx = context.WithValue(ctx, "dataDir", s.dataDir)
		ctx = context.WithValue(ctx, "thumbDir", s.thumbDir)
		ctx = context.WithValue(ctx, "badger", s.db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) Photos() http.Handler {
	r := chi.NewRouter()

	r.Get("/", s.PhotoHandler.List)
	r.Get("/*", s.PhotoHandler.Get)

	return r
}

func (s *server) Typeahead() http.Handler {
	r := chi.NewRouter()
	r.Route("/{source}", func(rs chi.Router) {
		rs.Get("/field", AutoCompleteField)
		rs.Get("/value", AutoCompleteValue)
	})
	r.Get("/name", AutoCompleteName)
	return r
}

func (s *server) Queries() http.Handler {
	r := chi.NewRouter()

	r.Get("/locations", QueryLocations)
	r.Get("/labels", QueryColorLabels)
	r.Get("/tags", QueryTags)
	r.Get("/faces", QueryFaces)
	r.Get("/rating", QueryRating)

	return r
}
