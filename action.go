package realm

import "github.com/Brickchain/go-document.v2"

type ControllerAction struct {
	document.ActionDescriptor
	ControllerID string `json:"controllerId,omitempty"` // belongs to controller
	OwnedByRealm bool   `json:"ownedByRealm,omitempty"` // if not provided by controller
	Signed       string `json:"signed,omitempty"`
}

type ActionProvider interface {
	List(realmID string) ([]*ControllerAction, error)
	Get(realmID, id string) (*ControllerAction, error)
	Set(realmID string, action *ControllerAction) error
	Delete(realmID, id string) error
	ListForController(realmID, controllerID string) ([]*ControllerAction, error)
}
