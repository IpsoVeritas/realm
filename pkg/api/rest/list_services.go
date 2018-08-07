package rest

import (
	"net/http"

	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
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

	total := stats.StartTimer("api.list_services.ListServices.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	req.Log().Debugf("Mandates before filtering: %+v", req.Mandates())

	mandates := make([]*document.Mandate, 0)
	for _, m := range context.MandatesForRealm(req.Mandates()) {
		mandates = append(mandates, m.Mandate)
	}

	req.Log().Debugf("Mandates after filtering: %+v", mandates)

	mp, err := context.Actions().Services(mandates)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list services"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, mp)
}
