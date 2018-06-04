package document

import (
	"time"

	jose "gopkg.in/square/go-jose.v1"
)

const ControllerDescriptorType = "controller-descriptor"

type KeyPurpose struct {
	DocumentType string `json:"documentType"`
	Required     bool   `json:"required,omitempty"`
	Description  string `json:"description,omitempty"`
}

type ControllerDescriptor struct {
	Base
	Label              string           `json:"label"`
	ActionsURI         string           `json:"actionsURI"`
	AdminUI            string           `json:"adminUI,omitempty"`
	BindURI            string           `json:"bindURI,omitempty"`
	Key                *jose.JsonWebKey `json:"key,omitempty"`
	KeyPurposes        []KeyPurpose     `json:"keyPurposes,omitempty"`
	RequireSetup       bool             `json:"requireSetup"`
	AddBindingEndpoint string           `json:"addBindingEndpoint,omitempty"`
	Icon               string           `json:"icon,omitempty"`
}

func NewControllerDescriptor(label string, actionsURI string) *ControllerDescriptor {
	return &ControllerDescriptor{
		Base: Base{
			Context:   Context,
			Type:      ControllerDescriptorType,
			Timestamp: time.Now(),
		},
		Label:      label,
		ActionsURI: actionsURI,
	}
}
