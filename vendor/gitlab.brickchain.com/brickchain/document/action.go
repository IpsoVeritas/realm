package document

import "time"

const ActionType = "action"

type Action struct {
	Base
	Role     string            `json:"role"`
	Mandate  string            `json:"mandate,omitempty"`
	Mandates []string          `json:"mandates,omitempty"`
	Nonce    string            `json:"nonce,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Facts    []Part            `json:"facts,omitempty"`
	Contract *Contract         `json:"contract,omitempty"`
}

func NewAction(realmRole string, mandate string) *Action {
	return &Action{
		Base: Base{
			Context:   Context,
			Type:      ActionType,
			Timestamp: time.Now(),
		},
		Role:     realmRole,
		Mandate:  mandate,
		Mandates: []string{mandate},
		Params:   make(map[string]string),
	}
}
