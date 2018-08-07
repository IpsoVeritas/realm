package services

import (
	"encoding/json"
	"time"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	logger "github.com/Brickchain/go-logger.v1"
	"github.com/pkg/errors"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	jose "gopkg.in/square/go-jose.v1"
)

type RealmService struct {
	base    string
	p       *RealmsServiceProvider
	realmID string
	realm   *realm.Realm
}

func NewRealmService(base string, p *RealmsServiceProvider, realmID string) *RealmService {
	return &RealmService{
		base:    base,
		p:       p,
		realmID: realmID,
	}
}

func (r *RealmService) Realm() (*realm.Realm, error) {
	var err error
	if r.realm == nil {
		if r.realm, err = r.p.realms.Get(r.realmID); err != nil {
			return nil, err
		}
	}

	return r.realm, nil
}

func (r *RealmService) Set(realm *realm.Realm) error {
	var err error
	realm.SignedDescriptor, err = r.p.signDescriptor(realm)
	if err != nil {
		return err
	}

	if err := r.p.realms.Set(realm); err != nil {
		return err
	}

	return r.p.publishEvent("UPDATED", r.realmID)
}

func (r *RealmService) Delete() error {
	actions, err := r.Actions().List()
	if err == nil {
		for _, action := range actions {
			r.Actions().Delete(action.ID)
		}
	}

	controllers, err := r.Controllers().List()
	if err == nil {
		for _, controller := range controllers {
			r.Controllers().Delete(controller.ID)
		}
	}

	mandates, err := r.Mandates().List()
	if err == nil {
		for _, mandate := range mandates {
			r.Mandates().Delete(mandate.ID)
		}
	}

	roles, err := r.Roles().List()
	if err == nil {
		for _, role := range roles {
			r.Roles().Delete(role.ID)
		}
	}

	if err := r.p.realms.Delete(r.realmID); err != nil {
		return err
	}

	return r.p.publishEvent("DELETED", r.realmID)
}

func (r *RealmService) Sign(payload []byte) (*jose.JsonWebSignature, error) {
	return r.p.signPayload(r.realmID, payload)
}

func (r *RealmService) Key() (*jose.JsonWebKey, error) {
	return r.p.getKey(r.realmID)
}

func (r *RealmService) HasMandateForRealm(mandates []httphandler.AuthenticatedMandate) bool {
	return r.HasAdminMandateForRealm(mandates)
}

func (r *RealmService) MandatesForRealm(mandates []httphandler.AuthenticatedMandate) []httphandler.AuthenticatedMandate {
	out := make([]httphandler.AuthenticatedMandate, 0)

	realm, err := r.Realm()
	if err != nil {
		return out
	}

	// make sure we never trigger the case where everything fails if the bootstrap realm hasn't been loaded
	bootstrapRealmTP := "something that should never match a thumbprint..."
	if r.p.bootstrapRealm != nil {
		bootstrapRealmTP = crypto.Thumbprint(r.p.bootstrapRealm.PublicKey)
	}

	realmTP := crypto.Thumbprint(realm.PublicKey)
	for _, m := range mandates {
		signerTP := crypto.Thumbprint(m.Signer)
		switch signerTP {
		case realmTP:
			if m.Mandate.Realm == r.realmID {
				out = append(out, m)
			} else {
				logger.Debugf("Mandate realm does not match context: %s != %s", m.Mandate.Realm, r.realmID)
			}
		case bootstrapRealmTP:
			if m.Mandate.Realm == r.p.bootstrapRealm.ID {
				out = append(out, m)
			}
		default:
			logger.Debugf("Mandate for role %s not signed by realm key. %s != %s", m.Mandate.Role, signerTP, realmTP)
		}
	}

	return out
}

func (r *RealmService) HasAdminMandateForRealm(mandates []httphandler.AuthenticatedMandate) bool {
	realm, err := r.Realm()
	if err != nil {
		return false
	}

	realmMandates := r.MandatesForRealm(mandates)

	for _, m := range realmMandates {
		for _, role := range realm.AdminRoles {
			if m.Mandate.Role == role {
				return true
			}
		}

		for _, role := range r.p.bootstrapRealm.AdminRoles {
			if m.Mandate.Role == role {
				return true
			}
		}
	}

	return false
}

func (r *RealmService) Actions() *ActionService {
	return &ActionService{
		base:             r.base,
		bootstrapRealmID: r.p.bootstrapRealmID,
		p:                r.p.actions,
		realmID:          r.realmID,
		realm:            r,
	}
}

func (r *RealmService) Controllers() *ControllerService {
	return &ControllerService{
		p:       r.p.controllers,
		realmID: r.realmID,
		realm:   r,
	}
}

func (r *RealmService) Invites() *InviteService {
	return &InviteService{
		base:    r.base,
		p:       r.p.invites,
		realmID: r.realmID,
		realm:   r,
		email:   r.p.email,
		assets:  r.p.assets,
	}
}

func (r *RealmService) Mandates() *MandateService {
	return &MandateService{
		p:            r.p.mandates,
		realmID:      r.realmID,
		realmContext: r,
	}
}

func (r *RealmService) MandateTickets() *MandateTicketService {
	return &MandateTicketService{
		p:       r.p.mandateTickets,
		realmID: r.realmID,
	}
}

func (r *RealmService) Roles() *RoleService {
	return &RoleService{
		p:       r.p.roles,
		realmID: r.realmID,
	}
}

func (r *RealmService) Settings() *SettingService {
	return &SettingService{
		p:       r.p.settings,
		realmID: r.realmID,
	}
}

func (r *RealmService) Files() *FileService {
	return &FileService{
		p:       r.p.filestore,
		realmID: r.realmID,
	}
}

func (r *RealmService) Join(jws *jose.JsonWebSignature) (*document.Multipart, error) {

	var userKey *jose.JsonWebKey
	var action document.Action
	var certificate *document.Certificate

	realm, err := r.Realm()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get realm")
	}

	if realm.GuestRole == "" {
		return nil, errors.New("No public role set")
	}

	if len(jws.Signatures) < 1 {
		return nil, errors.Wrap(err, "no key in signature")
	}

	actionBytes, err := jws.Verify(jws.Signatures[0].Header.JsonWebKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify signature")
	}

	if err := json.Unmarshal(actionBytes, &action); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal scope-request")
	}

	if action.Certificate != "" {
		certificate, err = crypto.VerifyCertificate(action.Certificate, 1000)
		if err != nil {
			return nil, errors.Wrap(err, "failed to verify certificate")
		}
	}

	if certificate != nil {
		userKey = certificate.Issuer
	} else {
		userKey = jws.Signatures[0].Header.JsonWebKey
	}

	role, err := r.Roles().ByName(realm.GuestRole)
	if err != nil {
		return nil, errors.Wrap(err, "could not get role")
	}

	mandate := document.NewMandate(realm.GuestRole)
	mandate.Realm = r.realmID
	mandate.RoleName = role.Description

	now := time.Now().UTC()
	mandate.ValidFrom = &now
	mandate.Recipient = userKey
	mandate.Sender = "realm"

	issued, err := r.Mandates().Issue(mandate, crypto.Thumbprint(userKey))
	if err != nil {
		return nil, errors.Wrap(err, "could not issue mandate")
	}

	part := document.Part{
		Encoding: "application/json+jws",
		Name:     "mandate",
		Document: issued.Signed,
	}
	multipart := document.NewMultipart()
	multipart.Append(part)

	return multipart, nil
}
