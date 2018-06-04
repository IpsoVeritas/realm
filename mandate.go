package realm

import "github.com/Brickchain/go-document.v2"

type IssuedMandate struct {
	document.Mandate
	Label  string `json:"label,omitempty"`
	Status int    `json:"status,omitempty"`
	Signed string `json:"signed,omitempty"`
}

type IssuedMandateProvider interface {
	List(realmID string) ([]*IssuedMandate, error)
	Get(realmID, id string) (*IssuedMandate, error)
	Set(realmID string, mandate *IssuedMandate) error
	Delete(realmID, id string) error
	ListForRole(realmID string, role string) ([]*IssuedMandate, error)
}
