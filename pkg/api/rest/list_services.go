package rest

import (
	"net/http"

	"github.com/IpsoVeritas/document"
	httphandler "github.com/IpsoVeritas/httphandler"
	"github.com/IpsoVeritas/realm/pkg/services"
	"github.com/pkg/errors"
)

type ServicesController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewServicesController(contextProvider *services.RealmsServiceProvider) *ServicesController {
	return &ServicesController{
		contextProvider: contextProvider,
	}
}

func (c *ServicesController) ListServices(req httphandler.OptionalAuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	mandates := make([]*document.Mandate, 0)
	for _, m := range context.MandatesForRealm(req.Mandates()) {
		mandates = append(mandates, m.Mandate)
	}

	mp, err := context.Actions().Services(mandates)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list services"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, mp)
}
