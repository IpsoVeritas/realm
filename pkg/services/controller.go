package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	resty "gopkg.in/resty.v0"
	jose "gopkg.in/square/go-jose.v1"
)

type ControllerService struct {
	p       realm.ControllerProvider
	realmID string
	realm   *RealmService
}

func (c *ControllerService) List() ([]*realm.Controller, error) {
	return c.p.List(c.realmID)
}

func (c *ControllerService) Get(id string) (*realm.Controller, error) {
	return c.p.Get(c.realmID, id)
}

func (c *ControllerService) Set(controller *realm.Controller) error {
	controller.Realm = c.realmID
	return c.p.Set(c.realmID, controller)
}

func (c *ControllerService) Delete(id string) error {
	actions, err := c.realm.Actions().ListForController(id)
	if err != nil {
		return errors.Wrap(err, "failed to list actions for controller")
	}

	for _, action := range actions {
		if err := c.realm.Actions().Delete(action.ID); err != nil {
			return errors.Wrapf(err, "failed to delete action %s", action.ID)
		}
	}

	controller, err := c.Get(id)
	if err != nil {
		return errors.Wrap(err, "failed to get controller")
	}

	if err := c.realm.Mandates().Delete(controller.MandateID); err != nil {
		return errors.Wrap(err, "failed to delete mandate for controller")
	}

	return c.p.Delete(c.realmID, id)
}

func (c *ControllerService) Bind(controller *realm.Controller) (*jose.JsonWebSignature, error) {
	if controller.ID == "" {
		controller.ID = crypto.Sha256(controller.Descriptor.BindURI)
	}

	realmData, err := c.realm.Realm()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get realm data")
	}

	if len(controller.AdminRoles) < 1 {
		controller.AdminRoles = realmData.AdminRoles
	}

	bind := document.NewControllerBinding(realmData.Descriptor)
	bind.ID = controller.ID

	purposes := []string{}
	for _, purpose := range controller.Descriptor.KeyPurposes {
		purposes = append(purposes, purpose.DocumentType)
	}

	role, err := c.realm.Roles().ByName(controller.MandateRole)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get role %s", controller.MandateRole)
	}

	realmKey, err := c.realm.Key()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get realm key")
	}

	cert, err := crypto.CreateCertificate(realmKey,
		controller.Descriptor.Key, role.KeyLevel, purposes, 0, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	bind.ControllerCertificate = cert
	bind.AdminRoles = controller.AdminRoles

	// if controller.Cert != "" {
	// 	controller.CertHistory = append(controller.CertHistory, controller.Cert)
	// }
	// controller.Cert = cert

	if controller.MandateRole == "" {
		controller.MandateRole = fmt.Sprintf("service@%s", c.realmID)
	}

	mandate := document.NewMandate(controller.MandateRole)
	// mandate.Label = fmt.Sprintf("Service: %s", controller.Name)
	mandate.Recipient = controller.Descriptor.Key
	mandate.Realm = c.realmID

	issued, err := c.realm.Mandates().Issue(mandate, fmt.Sprintf("Service: %s", controller.Name))
	if err != nil {
		return nil, errors.Wrap(err, "failed to issue mandate")
	}

	bind.Mandates = []string{issued.Signed}

	controller.MandateID = issued.ID

	// call controller, might not be accessible from realm-service
	ping, err := resty.SetTimeout(time.Second * 1).R().Get(controller.URI)
	if err == nil && ping.StatusCode() == 200 {
		controller.Reachable = true
	}

	err = c.realm.Controllers().Set(controller)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save controller")
	}

	bytes, err := json.Marshal(bind)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json")
	}

	return c.realm.Sign(bytes)
}

func (c *ControllerService) UpdateActions(controllerID string, mp *document.Multipart) error {

	controller, err := c.Get(controllerID)
	if err != nil {
		return errors.Wrap(err, "failed to get controller")
	}

	list, err := c.realm.Actions().ListForController(controllerID)
	if err != nil {
		return errors.Wrap(err, "faled to list actions for controller")
	}

	for _, part := range mp.Parts {
		updated := false
		// descriptor.Certificate = ""
		jws, err := crypto.UnmarshalSignature([]byte(part.Document))
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal JWS")
		}

		payload, err := jws.Verify(controller.Descriptor.Key)
		if err != nil {
			return errors.Wrap(err, "failed to verify action signature")
		}

		descriptor := &document.ActionDescriptor{}
		if err := json.Unmarshal(payload, &descriptor); err != nil {
			return errors.Wrap(err, "failed to unmarshal action")
		}

		for i, action := range list {
			if action.ID == descriptor.ID {
				// remove id from list
				list = append(list[:i], list[i+1:]...)

				action.ActionDescriptor = *descriptor
				action.Signed = part.Document

				if err := c.realm.Actions().Set(action); err != nil {
					return err
				}

				updated = true
			}
		}

		if !updated {
			if descriptor.ID == "" {
				descriptor.ID = uuid.Must(uuid.NewV4()).String()
			}
			action := &realm.ControllerAction{
				ActionDescriptor: *descriptor,
				ControllerID:     controllerID,
			}

			action.Signed = part.Document

			if err := c.realm.Actions().Set(action); err != nil {
				return err
			}
		}
	}
	if len(list) > 0 {
		for _, action := range list {
			if !action.OwnedByRealm {
				if err = c.realm.Actions().Delete(action.ID); err != nil {
					continue
				}
			}
		}
	}

	return nil
}
