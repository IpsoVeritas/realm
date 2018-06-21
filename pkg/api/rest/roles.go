package rest

import (
	"encoding/json"
	"net/http"

	document "github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
)

type RolesController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewRolesController(contextProvider *services.RealmsServiceProvider) *RolesController {
	return &RolesController{
		contextProvider: contextProvider,
	}
}

func (c *RolesController) ListRoles(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("handlers.roles.ListRoles.total")
	defer total.Stop()

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

func (c *RolesController) GetRole(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("handlers.roles.GetRole.total")
	defer total.Stop()

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

func (c *RolesController) SetRole(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("handlers.roles.Set.total")
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

	role := &document.Role{}
	if err := json.Unmarshal(body, &role); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal role"))
	}

	if err := context.Roles().Set(role); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrapf(err, "failed to store role"))
	}

	/*
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
	*/

	return httphandler.NewJsonResponse(http.StatusOK, role)
}

// func (c *RolesController) PostRole(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext) error {

// 	total := stats.StartTimer("handlers.roles.PostRole.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	data, err := readBody(r)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	m := &model.Role{}
// 	err = json.Unmarshal(data, &m)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	if m.KeyLevel < 10 {
// 		m.KeyLevel = 1000
// 	}

// 	id, err := realm.Roles().Post(m)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("got id: " + id)
// 	http.Redirect(w, r, r.URL.Path+"/"+id, http.StatusCreated)
// 	return nil
// }

// func (c *RolesController) UpdateRole(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext) error {

// 	total := stats.StartTimer("handlers.roles.UpdateRole.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	roleID := p.ByName("id")

// 	data, err := readBody(r)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	m := &model.Role{}
// 	err = json.Unmarshal(data, &m)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	if m.KeyLevel < 10 {
// 		m.KeyLevel = 1000
// 	}

// 	err = realm.Roles().Put(roleID, m)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("updated Role id: " + roleID)
// 	http.Redirect(w, r, r.URL.Path+"/"+roleID, http.StatusCreated)
// 	return nil
// }

// func (c *RolesController) DeleteRole(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext) error {

// 	total := stats.StartTimer("handlers.roles.DeleteRole.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	roleID := p.ByName("id")

// 	role, err := realm.Roles().Load(roleID)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	invites, err := realm.Invites().ListForRole(role.Name)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	for _, inviteID := range invites {
// 		err = realm.Invites().Delete(inviteID)
// 		if err != nil {
// 			log.Error(err)
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return nil
// 		}
// 	}

// 	mandates, err := realm.Mandates().ListForRole(role.Name)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	for _, mandateID := range mandates {
// 		mandate, err := realm.Mandates().Load(mandateID)
// 		if err != nil {
// 			log.Error(err)
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return nil
// 		}

// 		if err = mandate.Revoke(); err != nil {
// 			log.Error(err)
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return nil
// 		}
// 	}

// 	err = realm.Roles().Delete(roleID)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusNoContent)
// 	log.Infof("Role %s deleted from realm %s",
// 		roleID,
// 		realm.ID())

// 	return nil
// }
