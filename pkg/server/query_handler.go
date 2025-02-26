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

func AutoCompleteName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("q"))
	if name == "" {
		writeJSON(w, r, map[string]interface{}{
			"total":   0,
			"results": []string{},
		})
		return
	}

	results, err := photos.GetNames(r.Context(), name)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, r, results)
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
	if len(ls) == 0 {
		writeJSON(w, r, map[string][]string{
			"photos": []string{},
		})
		return
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
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	ps := PhotoSort(r.URL.Query().Get("sort"), r.URL.Query().Get("sort_dir") != "desc", count, offset, p)

	writeJSON(w, r, map[string]interface{}{
		"photos": ps,
	})
}

func QueryTags(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	tags := make([]string, 0, 5)
	for _, t := range strings.Split(r.URL.Query().Get("t"), ",") {
		if t != "" {
			tags = append(tags, t)
		}
	}

	if len(tags) == 0 {
		writeJSON(w, r, map[string][]string{
			"photos": []string{},
		})
		return
	}

	p, err := photos.ByTags(r.Context(), tags)
	if err != nil {
		log.WithError(err).Error("error getting photos by tags")
		writeErr(w, 500, err)
		return
	}

	count := 30
	if c, err := strconv.Atoi(r.URL.Query().Get("count")); err == nil {
		count = c
	}
	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	ps := PhotoSort(r.URL.Query().Get("sort"), r.URL.Query().Get("sort_dir") != "desc", count, offset, p)

	writeJSON(w, r, map[string]interface{}{
		"photos": ps,
	})

}

func QueryFaces(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)

	p, err := photos.HasFace(r.Context())
	if err != nil {
		log.WithError(err).Error("error getting photos with faces")
		writeErr(w, 500, err)
		return
	}

	count := 30
	if c, err := strconv.Atoi(r.URL.Query().Get("count")); err == nil {
		count = c
	}
	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	ps := PhotoSort(r.URL.Query().Get("sort"), r.URL.Query().Get("sort_dir") != "desc", count, offset, p)

	writeJSON(w, r, map[string]interface{}{
		"photos": ps,
	})

}

func QueryRating(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLog(r)
	rs := make([]string, 0, 3)
	for _, s := range strings.Split(r.URL.Query().Get("r"), ",") {
		if x, err := strconv.Atoi(s); err == nil && x <= 5 && x > -2 {
			rs = append(rs, s)
		}
	}
	if len(rs) == 0 {
		writeJSON(w, r, map[string][]string{
			"photos": []string{},
		})
		return
	}

	p, err := photos.ByRating(r.Context(), rs)
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
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	ps := PhotoSort(r.URL.Query().Get("sort"), r.URL.Query().Get("sort_dir") != "desc", count, offset, p)

	writeJSON(w, r, map[string]interface{}{
		"photos": ps,
	})
}
