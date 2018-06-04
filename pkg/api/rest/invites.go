package rest

import (
	"encoding/json"
	"net/http"

	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	"gitlab.brickchain.com/brickchain/crypto"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
)

type InvitesController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewInvitesController(contextProvider *services.RealmsServiceProvider) *InvitesController {
	return &InvitesController{
		contextProvider: contextProvider,
	}
}

func (c *InvitesController) List(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("handlers.invites.List.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	var err error
	var invites []*realm.Invite

	roleName := req.Params().ByName("roleName")
	if roleName != "" {
		invites, err = context.Invites().ListForRole(roleName)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not list invites for role"))
		}
	} else {
		invites, err = context.Invites().List()
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list invites"))
		}
	}

	return httphandler.NewJsonResponse(http.StatusOK, invites)
}

func (c *InvitesController) Get(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("handlers.invites.Get.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	inviteID := req.Params().ByName("inviteID")
	if inviteID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify invite ID"))
	}

	invite, err := context.Invites().Get(inviteID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get invite"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, invite)
}

func (c *InvitesController) Set(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("handlers.invites.New.total")
	defer total.Stop()

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

	invite := &realm.Invite{}
	if err := json.Unmarshal(body, &invite); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal invite"))
	}

	action := "create"
	inviteID := req.Params().ByName("inviteID")
	if inviteID != "" {
		if inviteID != invite.ID {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "tried to update invite with other ID than in payload"))
		}
		action = "update"
	}

	if err := context.Invites().Set(invite); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrapf(err, "failed to %s invite", action))
	}

	return httphandler.NewJsonResponse(http.StatusOK, invite)
}

func (c *InvitesController) Delete(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("handlers.invites.Delete.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	inviteID := req.Params().ByName("inviteID")
	if inviteID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify invite ID"))
	}

	if err := context.Invites().Delete(inviteID); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete invite"))
	}

	return httphandler.NewEmptyResponse(http.StatusNoContent)
}

func (c *InvitesController) Send(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("handlers.invites.Send.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	inviteID := req.Params().ByName("inviteID")
	if inviteID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify invite ID"))
	}

	invite, err := context.Invites().Get(inviteID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get invite"))
	}

	status, err := context.Invites().Send(invite)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete invite"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, status)
}

func (c *InvitesController) Fetch(req httphandler.Request) httphandler.Response {
	total := stats.StartTimer("handlers.invites.Send.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	inviteID := req.Params().ByName("inviteID")
	if inviteID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify invite ID"))
	}

	jws, err := context.Invites().Fetch(inviteID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get invite"))
	}

	return httphandler.NewStandardResponse(http.StatusOK, "application/json", jws.FullSerialize())
}

func (c *InvitesController) Callback(req httphandler.Request) httphandler.Response {
	total := stats.StartTimer("handlers.invites.Send.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	inviteID := req.Params().ByName("inviteID")
	if inviteID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify invite ID"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to read request body"))
	}

	jws, err := crypto.UnmarshalSignature(body)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal JWS"))
	}

	mp, err := context.Invites().Callback(inviteID, jws)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to process invite callback"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, mp)
}
