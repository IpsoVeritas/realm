package rest

import (
	"encoding/json"
	"net/http"

	document "github.com/IpsoVeritas/document"
	httphandler "github.com/IpsoVeritas/httphandler"
	realm "github.com/IpsoVeritas/realm"
	"github.com/IpsoVeritas/realm/pkg/services"
	"github.com/pkg/errors"
)

type MandatesController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewMandatesController(contextProvider *services.RealmsServiceProvider) *MandatesController {
	return &MandatesController{
		contextProvider: contextProvider,
	}
}

func (c *MandatesController) List(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	var err error
	var mandates []*realm.IssuedMandate

	roleName := req.Params().ByName("roleName")
	if roleName != "" {
		mandates, err = context.Mandates().ListForRole(roleName)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not list mandates for role"))
		}
	} else {
		mandates, err = context.Mandates().List()
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list mandates"))
		}
	}

	return httphandler.NewJsonResponse(http.StatusOK, mandates)
}

func (c *MandatesController) Get(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	mandateID := req.Params().ByName("mandateID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	mandate, err := context.Mandates().Get(mandateID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not get mandates"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, mandate)
}

func (c *MandatesController) Revoke(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	mandateID := req.Params().ByName("mandateID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	mandate, err := context.Mandates().Get(mandateID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not get mandates"))
	}

	mandate, err = context.Mandates().Revoke(mandate)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not revoke mandate"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, mandate)
}

func (c *MandatesController) Issue(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		if !c.contextProvider.HasMandateForBootstrapRealm(req.Mandates()) {
			return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No access to issue mandates"))
		}
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body"))
	}

	mandate := &document.Mandate{}
	if err := json.Unmarshal(body, &mandate); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to unmarshal mandate json"))
	}

	issued, err := context.Mandates().Issue(mandate, mandate.Recipient.KeyID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not issue mandate"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, issued)
}
