package server

import (
	"net/http"

	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/photos"
)

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
