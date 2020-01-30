package server

import (
	"net/http"
	"strconv"
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

func QueryColorLabels(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	l := r.URL.Query().Get("l")
	ls := make([]string, 0, 4)
	for _, s := range strings.Split(l, ",") {
		if x, err := strconv.Atoi(s); err == nil && x <= 5 {
			ls = append(ls, s)
		}
	}

	p, err := photos.ColorLabels(r.Context(), ls)
	if err != nil {
		log.WithError(err).Error("error getting photos with color labels")
		writeErr(w, 500, err)
		return
	}

	count := 30
	if c, err := strconv.Atoi(r.URL.Query().Get("count")); err == nil {
		count = c
	}
	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("count")); err == nil {
		offset = o
	}
	ps := PhotoSort(r.URL.Query().Get("sort"), r.URL.Query().Get("sort_dir") != "desc", count, offset, p)

	writeJSON(w, r, map[string]interface{}{
		"photos": ps,
	})
}
