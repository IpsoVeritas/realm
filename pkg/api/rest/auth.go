package rest

import (
	httphandler "github.com/IpsoVeritas/httphandler"
	"github.com/pkg/errors"
	"github.com/IpsoVeritas/realm/pkg/services"

	"net/http"
)

type AuthController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewAuthController(contextProvider *services.RealmsServiceProvider) *AuthController {
	return &AuthController{
		contextProvider: contextProvider,
	}
}

func (c *AuthController) Authenticated(req httphandler.AuthenticatedRequest) httphandler.Response {

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	resp := struct {
		Authenticated bool `json:"authenticated"`
	}{
		Authenticated: true,
	}

	return httphandler.NewJsonResponse(http.StatusOK, resp)
}
