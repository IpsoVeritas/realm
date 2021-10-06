package realm

import (
	"github.com/IpsoVeritas/document"
	uuid "github.com/satori/go.uuid"
)

type MandateTicket struct {
	document.Base
	Mandate      *document.Mandate      `json:"mandate,omitempty"`
	ScopeRequest *document.ScopeRequest `json:"scope-request,omitempty"`
	Facts        map[string]string      `json:"facts,omitempty"`
	Static       bool                   `json:"static,omitempty"`
}

func NewMandateTicket() *MandateTicket {
	return &MandateTicket{
		Base: document.Base{
			ID: uuid.NewV4().String(),
		},
		Facts: make(map[string]string),
	}
}

type MandateTicketProvider interface {
	List(realmID string) ([]*MandateTicket, error)
	Get(realmID, id string) (*MandateTicket, error)
	Set(realmID string, ticket *MandateTicket) error
	Delete(realmID, id string) error
}
