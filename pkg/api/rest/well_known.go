package rest

import (
	"net/http"

	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	"github.com/Brickchain/realm/pkg/services"
)

type WellKnownHandler struct {
	base            string
	contextProvider *services.RealmsServiceProvider
}

func NewWellKnownHandler(base string, contextProvider *services.RealmsServiceProvider) *WellKnownHandler {
	return &WellKnownHandler{
		base:            base,
		contextProvider: contextProvider,
	}
}

func (h *WellKnownHandler) WellKnown(req httphandler.Request) httphandler.Response {
	total := stats.StartTimer("api.well_known.WellKnown.total")
	defer total.Stop()

	realmID := req.OriginalRequest().Host

	req.Log().AddField("realm", realmID)

	context := h.contextProvider.Get(realmID)

	realm, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusNotFound, errors.Wrap(err, "realm not found"))
	}

	r := httphandler.NewStandardResponse(http.StatusOK, "application/json", realm.SignedDescriptor)
	r.Header().Set("Cache-Control", "max-age=300")

	return r
}

func (h *WellKnownHandler) WellKnownForRealm(req httphandler.Request) httphandler.Response {
	total := stats.StartTimer("api.well_known.WellKnownForRealm.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")

	req.Log().AddField("realm", realmID)

	context := h.contextProvider.Get(realmID)

	realm, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusNotFound, errors.Wrap(err, "realm not found"))
	}

	r := httphandler.NewStandardResponse(http.StatusOK, "application/json", realm.SignedDescriptor)
	r.Header().Set("Cache-Control", "max-age=300")

	return r
}
