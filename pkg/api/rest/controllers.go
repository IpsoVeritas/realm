package rest

import (
	"encoding/json"
	"net/http"

	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	realm "github.com/Brickchain/realm"
	"github.com/Brickchain/realm/pkg/services"
)

type ControllersController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewControllersController(contextProvider *services.RealmsServiceProvider) *ControllersController {

	return &ControllersController{
		contextProvider: contextProvider,
	}
}

func (c *ControllersController) ListControllers(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("api.controllers.ListControllers.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	list, err := context.Controllers().List()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list controllers"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, list)

}

func (c *ControllersController) GetController(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("handlers.controllers.GetController.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	controllerID := req.Params().ByName("controllerID")
	if controllerID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify controller"))
	}

	controller, err := context.Controllers().Get(controllerID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get controller"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, controller)
}

func (c *ControllersController) Set(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("api.controllers.Set.total")
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

	controller := &realm.Controller{}
	if err := json.Unmarshal(body, &controller); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal controller"))
	}

	action := "create"
	controllerID := req.Params().ByName("controllerID")
	if controllerID != "" {
		if controllerID != controller.ID {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "tried to update controller with other ID than in payload"))
		}
		action = "update"
	}

	if err := context.Controllers().Set(controller); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrapf(err, "failed to %s controller", action))
	}

	return httphandler.NewJsonResponse(http.StatusOK, controller)
}

func (c *ControllersController) Bind(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("api.controllers.Bind.total")
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

	controller := &realm.Controller{}
	if err := json.Unmarshal(body, &controller); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal controller"))
	}

	jws, err := context.Controllers().Bind(controller)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to bind controller"))
	}

	return httphandler.NewStandardResponse(http.StatusOK, "application/json", jws.FullSerialize())
}

func (c *ControllersController) Delete(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("api.controllers.Delete.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	controllerID := req.Params().ByName("controllerID")
	if controllerID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify controller ID"))
	}

	if err := context.Controllers().Delete(controllerID); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete controller"))
	}

	return httphandler.NewEmptyResponse(http.StatusNoContent)
}

func (c *ControllersController) UpdateActions(req httphandler.AuthenticatedRequest) httphandler.Response {
	total := stats.StartTimer("api.controllers.UpdateActions.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	controllerID := req.Params().ByName("controllerID")
	if controllerID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify controller ID"))
	}

	// controller, err := context.Controllers().Get(controllerID)
	// if err != nil {
	// 	return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not load controller"))
	// }

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "could not read request body"))
	}

	mp := &document.Multipart{}
	if err := json.Unmarshal(body, &mp); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal multipart"))
	}

	// jws, err := crypto.UnmarshalSignature(body)
	// if err != nil {
	// 	return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal JWS"))
	// }

	// payload, err := jws.Verify(controller.Descriptor.Key)
	// if err != nil {
	// 	return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "could not verify signature"))
	// }

	// var descriptors []*document.ActionDescriptor
	// if err := json.Unmarshal(payload, &descriptors); err != nil {
	// 	return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal document"))
	// }

	if err := context.Controllers().UpdateActions(controllerID, mp, req.Key()); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to update actions"))
	}

	return httphandler.NewEmptyResponse(http.StatusCreated)
}

// func (c *ControllersController) PostController(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params) error {

// 	total := stats.StartTimer("handlers.controllers.PostController.total")
// 	defer total.Stop()

// 	realm, err := c.realmProvider.Get(p.ByName("rid"))
// 	if err != nil {
// 		logger.Errorf("error posting controller to realm %s, err: %s", realm.ID(), err)
// 		return err
// 	}

// 	data, err := readBody(r)

// 	if err != nil {
// 		logger.Errorf("error reading post data %s", err)
// 		return err
// 	}

// 	ctrl := &model.Controller{}
// 	err = json.Unmarshal(data, &ctrl)
// 	if err != nil {
// 		logger.Errorf("error reading input / unmarshal %s error: %s", data, err)
// 		return err
// 	}

// 	ctrl.Type = "controller"
// 	id, err := realm.Controllers().Post(ctrl)

// 	if err != nil {
// 		logger.Errorf("error posting data for %v", ctrl)
// 		return err
// 	}

// 	logger.Infof("got id: " + id)
// 	http.Redirect(w, r, r.RequestURI+"/"+id, http.StatusCreated)
// 	return nil

// }

// // uses crypto service and gives response to the controller back to client for forwarding

// type BindController struct {
// 	TTL      int      `json:"ttl,omitempty"`
// 	Purposes []string `json:"purposes,omitempty"`
// }

// func (c *ControllersController) BindController(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext) error {

// 	total := stats.StartTimer("handlers.controllers.BindController.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	body, err := readBody(r)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return nil
// 	}

// 	controller := &model.Controller{}
// 	err = json.Unmarshal(body, &controller)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return nil
// 	}

// 	if controller.ID == "" {
// 		controller.ID = crypto.Sha256(controller.Descriptor.BindURI)
// 	}

// 	if len(controller.AdminRoles) < 1 {
// 		controller.AdminRoles = realm.Realm().AdminRoles
// 	}

// 	bind := document.NewControllerBinding(realm.Realm().Descriptor)
// 	bind.ID = controller.ID

// 	// make sure endpoint is set on the descriptor
// 	if bind.RealmDescriptor.Endpoints == nil {
// 		bind.RealmDescriptor.Endpoints = map[string]string{
// 			"v1": fmt.Sprintf("%s/realm-api", viper.GetString("base")),
// 		}
// 	}

// 	purposes := []string{}
// 	for _, purpose := range controller.Descriptor.KeyPurposes {
// 		purposes = append(purposes, purpose.DocumentType)
// 	}

// 	logger.Infof("signing %s with %s, purposes: %v, ttl: %d",
// 		controller.Descriptor.Key.KeyID, realm.ID(), purposes, controller.TTL)

// 	info, err := c.crypto.InfoForKey(realm.ID())
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("loaded realm key: %v", info)

// 	role, err := realm.Roles().ByName(controller.MandateRole)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	cert, err := c.crypto.CreateCertificateChainWithKey(realm.ID(),
// 		controller.Descriptor.Key, role.KeyLevel, purposes, controller.TTL)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	bind.ControllerCertificateChain = cert
// 	bind.AdminRoles = controller.AdminRoles

// 	if controller.Cert != "" {
// 		controller.CertHistory = append(controller.CertHistory, controller.Cert)
// 	}
// 	controller.Cert = cert

// 	if controller.MandateRole == "" {
// 		controller.MandateRole = fmt.Sprintf("service@%s", realm.ID())
// 	}

// 	mandate := document.NewMandate(controller.MandateRole)
// 	mandate.Label = fmt.Sprintf("Service: %s", controller.Name)
// 	mandate.Recipient = controller.Descriptor.Key.KeyID
// 	mandate.RecipientName = fmt.Sprintf("Service: %s", controller.Name)
// 	mandate.RecipientPK = controller.Descriptor.Key
// 	mandate.Realm = realm.ID()
// 	mandate.TTL = controller.TTL

// 	issued, err := realm.Mandates().Issue(mandate)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	bind.Mandate, err = issued.CompactSigned()
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	controller.MandateID = issued.ID

// 	// call controller, might not be accessible from realm-service

// 	ping, err := resty.SetTimeout(time.Second * 1).R().Get(controller.URI)
// 	if err == nil && ping.StatusCode() == 200 {
// 		controller.Reachable = true
// 	} else {
// 		logger.Infof("Controller %s not reachable", controller.Name)
// 	}

// 	err = realm.Controllers().Put(controller.ID, controller)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("bound controller %s/%s to realm %s",
// 		controller.ID, controller.Name, realm.ID())

// 	bytes, err := json.Marshal(bind)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("bind: '%s'", string(bytes))

// 	signedBinding, err := c.crypto.SignWithKey(realm.ID(), string(bytes))
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.Path, controller.ID))
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	w.Write([]byte(signedBinding))

// 	return nil
// }

// func (c *ControllersController) DeleteController(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext,
// 	controller *model.Controller) error {

// 	total := stats.StartTimer("handlers.controllers.DeleteController.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	controllerID := p.ByName("id")

// 	if controller.MandateID != "" {
// 		mandate, err := realm.Mandates().Load(controller.MandateID)
// 		if err != nil {
// 			log.Error(err)
// 		}
// 		if err = mandate.Revoke(); err != nil {
// 			log.Error(err)
// 		}
// 	}

// 	actions, err := realm.Actions().List()
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	for _, actionID := range actions {
// 		action, err := realm.Actions().Load(actionID)
// 		if err != nil {
// 			log.Error(err)
// 			continue
// 		}

// 		if action.ControllerID == controllerID {
// 			realm.Actions().Delete(actionID)
// 		}
// 	}

// 	err = realm.Controllers().Delete(controllerID)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusNoContent)
// 	w.Write([]byte(""))
// 	log.Infof("deleted controller %s/%s from realm %s",
// 		controllerID, controller.Name, realm.ID)

// 	return nil
// }

// func (c *ControllersController) UpdateController(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext,
// 	controller *model.Controller) error {

// 	total := stats.StartTimer("handlers.controllers.UpdateController.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	controllerID := p.ByName("id")

// 	data, err := readBody(r)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	ctrl := &model.Controller{}
// 	err = json.Unmarshal(data, &ctrl)
// 	ctrl.Type = "controller"
// 	ctrl.ID = controllerID
// 	ctrl.Modified = time.Now().Unix()

// 	err = realm.Controllers().Put(controllerID, ctrl)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return nil
// 	}

// 	logger.Infof("updated controller.id: %s", controllerID)
// 	http.Redirect(w, r, r.RequestURI+"/"+controllerID,
// 		http.StatusCreated)
// 	return nil
// }

// func (c *ControllersController) PostActions(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext,
// 	controller *model.Controller) error {

// 	total := stats.StartTimer("handlers.controllers.PostActions.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	controllerID := p.ByName("id")

// 	data, err := readBody(r)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, "Could not read body", http.StatusInternalServerError)
// 		return nil
// 	}

// 	payload := data

// 	jws, err := crypto.UnmarshalSignature(data)
// 	if err == nil {
// 		payload, err = jws.Verify(controller.Descriptor.Key)
// 		if err != nil {
// 			log.Error(err)
// 			http.Error(w, "Could not verify signature", http.StatusInternalServerError)
// 			return nil
// 		}
// 	}

// 	log.Debug("Payload: ", string(payload))

// 	var descriptors []*document.ActionDescriptor
// 	err = json.Unmarshal(payload, &descriptors)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, "Could not unmarshal body", http.StatusInternalServerError)
// 		return nil
// 	}

// 	list, err := realm.Actions().ByController(controllerID)
// 	if err != nil {
// 		log.Error(err)
// 		return err
// 	}

// 	for _, descriptor := range descriptors {
// 		updated := false
// 		descriptor.CertificateChain = ""

// 		for i, actionID := range list {
// 			if actionID == descriptor.ID {
// 				// remove id from list
// 				list = append(list[:i], list[i+1:]...)

// 				action, err := realm.Actions().Load(actionID)
// 				if err != nil {
// 					log.Error(err)
// 					continue
// 				}
// 				log.Debugf("Updating action: %+v", action)
// 				action.ActionDescriptor = *descriptor

// 				if err = realm.Actions().Put(action.ID, action); err != nil {
// 					log.Error(err)
// 					http.Error(w, err.Error(), http.StatusInternalServerError)
// 					return nil
// 				}

// 				updated = true
// 			}
// 		}

// 		if !updated {
// 			if descriptor.ID == "" {
// 				descriptor.ID = uuid.Must(uuid.NewV4()).String()
// 			}
// 			action := &model.ControllerAction{
// 				ActionDescriptor: *descriptor,
// 				ControllerID:     controllerID,
// 			}
// 			err = realm.Actions().Put(action.ID, action)
// 			if err != nil {
// 				log.Error(err)
// 				http.Error(w, "Could not save action", http.StatusInternalServerError)
// 				return nil
// 			}

// 			log.Infof("Saved action with id: %s", action.ID)
// 		}
// 	}
// 	if len(list) > 0 {
// 		for _, id := range list {
// 			action, err := realm.Actions().Load(id)
// 			if err != nil {
// 				log.Error(err)
// 				continue
// 			}
// 			if !action.OwnedByRealm {
// 				log.Infof("Deleting action %s", id)
// 				if err = realm.Actions().Delete(id); err != nil {
// 					log.Error(err)
// 					continue
// 				}
// 			}
// 		}
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	return nil
// }

// func (c *ControllersController) GetActions(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	p httprouter.Params,
// 	user *auth.User,
// 	realm *model.RealmContext,
// 	controller *model.Controller) error {

// 	total := stats.StartTimer("handlers.controllers.GetActions.total")
// 	defer total.Stop()

// 	log := logger.ForContext(r.Context())

// 	controllerID := p.ByName("id")
// 	var descriptors []*document.ActionDescriptor
// 	list, err := realm.Actions().List()
// 	if err != nil {
// 		return err
// 	}
// 	for _, actionID := range list {
// 		action, err := realm.Actions().Load(actionID)
// 		if err != nil {
// 			log.Error(err)
// 			continue
// 		}
// 		if action.ControllerID == controllerID {
// 			descriptors = append(descriptors, &action.ActionDescriptor)
// 		}
// 	}

// 	return helper.SendPayload(w, descriptors, http.StatusOK)
// }
