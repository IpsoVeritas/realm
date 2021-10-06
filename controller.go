package realm

import "github.com/IpsoVeritas/document"

type Controller struct {
	document.Base
	Active       bool                           `json:"active"`
	Name         string                         `json:"name,omitempty"`
	URI          string                         `json:"uri,omitempty"`
	Descriptor   *document.ControllerDescriptor `json:"descriptor,omitempty"`
	Certificates []string                       `json:"certificates,omitempty"`
	AdminRoles   []string                       `json:"adminRoles,omitempty"`
	MandateID    string                         `json:"mandateID,omitempty"`
	Reachable    bool                           `json:"reachable,omitempty"`
	Hidden       bool                           `json:"hidden,omitempty"`
	Tags         []string                       `json:"tags,omitempty"`
	Priority     int                            `json:"priority,omitempty"`
	MandateRole  string                         `json:"mandateRole,omitempty"`
	ServiceID    string                         `json:"serviceID,omitempty"`
}

type ControllerProvider interface {
	List(realmID string) ([]*Controller, error)
	Get(realmID, id string) (*Controller, error)
	Set(realmID string, controller *Controller) error
	Delete(realmID, id string) error
}
