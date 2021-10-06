package rest

import (
	"net/http"

	httphandler "github.com/IpsoVeritas/httphandler"
	"github.com/IpsoVeritas/realm/pkg/services"
	"github.com/pkg/errors"
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
