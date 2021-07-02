package services

import (
	"encoding/json"

	"github.com/Brickchain/go-document.v2"

	crypto "github.com/Brickchain/go-crypto.v2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	realm "github.com/Brickchain/realm"
)

type MandateService struct {
	p            realm.IssuedMandateProvider
	realmID      string
	realmContext *RealmService
}

func (m *MandateService) List() ([]*realm.IssuedMandate, error) {
	return m.p.List(m.realmID)
}

func (m *MandateService) Get(id string) (*realm.IssuedMandate, error) {
	return m.p.Get(m.realmID, id)
}

func (m *MandateService) Set(mandate *realm.IssuedMandate) error {
	if mandate.ID == "" {
		mandate.ID = uuid.Must(uuid.NewV4()).String()
	}

	mandate.Realm = m.realmID
	return m.p.Set(m.realmID, mandate)
}

func (m *MandateService) Delete(id string) error {
	return m.p.Delete(m.realmID, id)
}

func (m *MandateService) ListForRole(role string) ([]*realm.IssuedMandate, error) {
	return m.p.ListForRole(m.realmID, role)
}

func (m *MandateService) Issue(mandate *document.Mandate, label string) (*realm.IssuedMandate, error) {
	if mandate.ID == "" {
		mandate.ID = uuid.Must(uuid.NewV4()).String()
	}
	mandate.Realm = m.realmID

	bytes, err := json.Marshal(mandate)
	if err != nil {
		return nil, err
	}

	jws, err := m.realmContext.Sign(bytes)
	if err != nil {
		return nil, err
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		return nil, err
	}

	issued := &realm.IssuedMandate{
		Label:   label,
		Mandate: *mandate,
		Signed:  compact,
	}

	if err := m.Set(issued); err != nil {
		return nil, err
	}

	return issued, nil
}

func (m *MandateService) Revoke(issued *realm.IssuedMandate) (*realm.IssuedMandate, error) {

	_, err := crypto.UnmarshalSignature([]byte(issued.Signed))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signature")
	}

	// TODO: Implement revocations logic

	issued.Status = document.MandateRevoked

	if err := m.Set(issued); err != nil {
		return nil, err
	}

	return issued, nil
}

// func (m *MandateService) Revoke(mandate *realm.IssuedMandate) error {
// 	sig, err := crypto.UnmarshalSignature([]byte(mandate.Signed))
// 	if err != nil {
// 		return errors.Wrap(err, "failed to unmarshal signature")
// 	}

// 	key, err := m.realmContext.Key()
// 	if err != nil {
// 		return err
// 	}

// 	revReq, err :=revocation.CreateRevocationRequest(key, sig, &sig.Signatures[0])
// 	if err != nil {
// 		return errors.Wrap(err, "Failed to create revocation request")
// 	}

// 	_, err = m.rev.SendRequest(revReq)
// 	if err != nil {
// 		return errors.Wrap(err, "Failed to send revocation")
// 	}

// 	m.Status = document.MandateRevoked
// }
