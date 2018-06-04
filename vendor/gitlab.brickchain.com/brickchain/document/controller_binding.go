package document

import (
	"time"
)

const ControllerBindingType = "controller-binding"

type ControllerBinding struct {
	Base
	RealmDescriptor            *RealmDescriptor `json:"realmDescriptor"`
	AdminRoles                 []string         `json:"adminRoles,omitempty"`
	ControllerCertificateChain string           `json:"controllerCertificateChain,omitempty"`
	Mandate                    string           `json:"mandate,omitempty"`
}

func NewControllerBinding(realmDescriptor *RealmDescriptor) *ControllerBinding {
	return &ControllerBinding{
		Base: Base{
			Context:   Context,
			Type:      ControllerBindingType,
			Timestamp: time.Now(),
		},
		RealmDescriptor: realmDescriptor,
	}
}
