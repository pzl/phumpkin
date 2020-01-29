package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/photos"
)

func AutoCompleteField(w http.ResponseWriter, r *http.Request) {
	var source byte
	sourceParam := chi.URLParam(r, "source")
	switch strings.ToLower(sourceParam) {
	case "xmp":
		source = photos.SourceXMP
	case "exif":
		source = photos.SourceEXIF
	default:
		writeFail(w, 400, "invalid source")
		return
	}

	keys, err := photos.GetFields(r.Context(), source, r.URL.Query().Get("q"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, r, map[string]interface{}{
		"keys": keys,
	})
}

func AutoCompleteValue(w http.ResponseWriter, r *http.Request) {
	var source byte
	sourceParam := chi.URLParam(r, "source")
	switch strings.ToLower(sourceParam) {
	case "xmp":
		source = photos.SourceXMP
	case "exif":
		source = photos.SourceEXIF
	default:
		writeFail(w, 400, "invalid source")
		return
	}

	values, err := photos.GetValues(r.Context(), source, r.URL.Query().Get("field"), r.URL.Query().Get("q"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, r, map[string]interface{}{
		"values": values,
	})
}

func QueryLocations(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)

	p, err := photos.WithLocation(r.Context())
	if err != nil {
		log.WithError(err).Error("error getting photos with locations")
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, r, map[string]interface{}{
		"photos": p,
	})
}
