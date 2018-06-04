package document

import "time"

const ActionDescriptorType = "action-descriptor"

type ActionDescriptor struct {
	Base
	Binding   string            `json:"binding,omitempty"`
	Label     string            `json:"label"`
	Roles     []string          `json:"roles"`
	UIURI     string            `json:"uiURI,omitempty"`
	UIData    string            `json:"uiData,omitempty"`
	NonceType string            `json:"nonceType,omitempty"`
	Nonce     string            `json:"nonce,omitempty"`
	NonceURI  string            `json:"nonceURI,omitempty"`
	ActionURI string            `json:"actionURI"`
	Params    map[string]string `json:"params,omitempty"`
	Scopes    []Scope           `json:"scopes,omitempty"`
	Icon      string            `json:"icon,omitempty"`
	KeyLevel  int               `json:"keyLevel,omitempty"`
	Internal  bool              `json:"internal,omitempty"`
	Contract  *Contract         `json:"contract,omitempty"`
}

func NewActionDescriptor(label string, roles []string, keyLevel int, actionURI string) *ActionDescriptor {
	return &ActionDescriptor{
		Base: Base{
			Context:   Context,
			Type:      ActionDescriptorType,
			Timestamp: time.Now(),
		},
		Label:     label,
		Roles:     roles,
		ActionURI: actionURI,
		KeyLevel:  keyLevel,
	}
}
