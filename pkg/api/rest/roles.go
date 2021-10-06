package rest

import (
	"encoding/json"
	"net/http"

	document "github.com/IpsoVeritas/document"
	httphandler "github.com/IpsoVeritas/httphandler"
	"github.com/IpsoVeritas/realm/pkg/services"
	"github.com/pkg/errors"
)

type RolesController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewRolesController(contextProvider *services.RealmsServiceProvider) *RolesController {
	return &RolesController{
		contextProvider: contextProvider,
	}
}

func (c *RolesController) List(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	list, err := context.Roles().List()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list roles"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, list)
}

func (c *RolesController) Get(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	roleID := req.Params().ByName("roleID")
	if roleID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify role ID"))
	}

	role, err := context.Roles().Get(roleID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get role"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, role)
}

func (c *RolesController) Set(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to read request body"))
	}

	role := &document.Role{}
	if err := json.Unmarshal(body, &role); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal role"))
	}

	if role.KeyLevel < 10 {
		role.KeyLevel = 1000
	}

	if err := context.Roles().Set(role); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrapf(err, "failed to store role"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, role)
}

func (c *RolesController) Delete(req httphandler.AuthenticatedRequest) httphandler.Response {
	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	roleID := req.Params().ByName("roleID")
	if roleID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify role ID"))
	}

	role, err := context.Roles().Get(roleID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get role"))
	}

	invites, err := context.Invites().ListForRole(role.Name)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get invites"))
	}

	for _, invite := range invites {
		err = context.Invites().Delete(invite.ID)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete invite"))
		}
	}

	mandates, err := context.Mandates().ListForRole(role.Name)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get mandates"))
	}

	for _, mandate := range mandates {
		_, err = context.Mandates().Revoke(mandate)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to revoke mandate"))
		}
	}

	if err := context.Roles().Delete(roleID); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete role"))
	}

	return httphandler.NewEmptyResponse(http.StatusNoContent)
}
