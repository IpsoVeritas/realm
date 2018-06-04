package services

import (
	uuid "github.com/satori/go.uuid"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
)

type MandateTicketService struct {
	p            realm.MandateTicketProvider
	realmID      string
	realmContext *RealmService
}

func (m *MandateTicketService) List() ([]*realm.MandateTicket, error) {
	return m.p.List(m.realmID)
}

func (m *MandateTicketService) Get(id string) (*realm.MandateTicket, error) {
	return m.p.Get(m.realmID, id)
}

func (m *MandateTicketService) Set(mandateTicket *realm.MandateTicket) error {
	if mandateTicket.ID == "" {
		mandateTicket.ID = uuid.Must(uuid.NewV4()).String()
	}

	mandateTicket.Realm = m.realmID
	return m.p.Set(m.realmID, mandateTicket)
}

func (m *MandateTicketService) Delete(id string) error {
	return m.p.Delete(m.realmID, id)
}
