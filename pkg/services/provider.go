package services

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	filestore "gitlab.brickchain.com/libs/go-filestore.v1"
	keys "gitlab.brickchain.com/libs/go-keys.v1"
	"gitlab.brickchain.com/libs/go-pubsub.v1"
	jose "gopkg.in/square/go-jose.v1"
)

type RealmsServiceProvider struct {
	base                  string
	realms                realm.RealmProvider
	actions               realm.ActionProvider
	controllers           realm.ControllerProvider
	invites               realm.InviteProvider
	mandates              realm.IssuedMandateProvider
	mandateTickets        realm.MandateTicketProvider
	roles                 realm.RoleProvider
	settings              realm.SettingProvider
	filestore             filestore.Filestore
	sks                   keys.StoredKeyService
	kek                   []byte
	pubsub                pubsub.PubSubInterface
	realmTopic            string
	bootstrapRealmID      string
	bootstrapRealmContext *RealmService
	bootstrapRealm        *realm.Realm
	keyset                *jose.JsonWebKeySet
	email                 realm.EmailProvider
	assets                realm.AssetProvider
}

func NewRealmsServiceProvider(
	base string,
	realms realm.RealmProvider,
	actions realm.ActionProvider,
	controllers realm.ControllerProvider,
	invites realm.InviteProvider,
	mandates realm.IssuedMandateProvider,
	mandateTickets realm.MandateTicketProvider,
	roles realm.RoleProvider,
	settings realm.SettingProvider,
	filestore filestore.Filestore,
	sks keys.StoredKeyService,
	kek []byte,
	ps pubsub.PubSubInterface,
	realmTopic string,
	keyset *jose.JsonWebKeySet,
	email realm.EmailProvider,
	assets realm.AssetProvider,
) *RealmsServiceProvider {

	r := &RealmsServiceProvider{
		base:           base,
		realms:         realms,
		actions:        actions,
		controllers:    controllers,
		invites:        invites,
		mandates:       mandates,
		mandateTickets: mandateTickets,
		roles:          roles,
		settings:       settings,
		filestore:      filestore,
		sks:            sks,
		kek:            kek,
		pubsub:         ps,
		realmTopic:     realmTopic,
		keyset:         keyset,
		email:          email,
		assets:         assets,
	}

	return r
}

func (p *RealmsServiceProvider) SetBase(base string) {
	p.base = base
}

func (p *RealmsServiceProvider) LoadBootstrapRealm(bootstrapRealmID string) error {
	p.bootstrapRealmID = bootstrapRealmID
	p.bootstrapRealmContext = p.Get(bootstrapRealmID)

	var err error
	if p.bootstrapRealm, err = p.bootstrapRealmContext.Realm(); err != nil {
		return errors.Wrap(err, "failed to load bootstrap realm")
	}

	return nil
}

func (p *RealmsServiceProvider) Bootstrap(password string) (*realm.MandateTicket, error) {
	if p.bootstrapRealm == nil {
		return nil, errors.New("Bootstrap realm not loaded")
	}

	bootstrapped, err := p.bootstrapRealmContext.Settings().Get("bootstrapped")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read bootstrap status")
	}

	if bootstrapped == "true" {
		return nil, errors.New("Already bootstrapped")
	}

	bootPW, err := p.bootstrapRealmContext.Settings().Get("password")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bootstrap password")
	}

	if password != bootPW {
		return nil, errors.New("Wrong password for bootstrapping realm")
	}

	if len(p.bootstrapRealm.AdminRoles) < 1 {
		return nil, errors.New("No admin roles for realm")
	}

	ticket := realm.NewMandateTicket()
	ticket.Realm = p.bootstrapRealm.ID

	ticket.Mandate = document.NewMandate(p.bootstrapRealm.AdminRoles[0])

	ticket.ScopeRequest = document.NewScopeRequest(1000)
	ticket.ScopeRequest.Contract = document.NewContract()
	ticket.ScopeRequest.Contract.Text = fmt.Sprintf("Become bootstrap admin for %s", ticket.Realm)
	ticket.ScopeRequest.Scopes = []document.Scope{
		{
			Name:     "https://schema.brickchain.com/v2/fact.json#name",
			Required: true,
		},
	}

	if err := p.mandateTickets.Set(p.bootstrapRealm.ID, ticket); err != nil {
		return nil, err
	}

	if err := p.bootstrapRealmContext.Settings().Set("bootstrapped", "true"); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (p *RealmsServiceProvider) HasMandateForBootstrapRealm(mandates []httphandler.AuthenticatedMandate) bool {
	bootstrapRealmTP := crypto.Thumbprint(p.bootstrapRealm.PublicKey)
	for _, m := range mandates {
		signerTP := crypto.Thumbprint(m.Signer)
		if signerTP == bootstrapRealmTP {
			if m.Mandate.Realm == p.bootstrapRealm.ID {
				return true
			}
		}
	}

	return false
}

func (p *RealmsServiceProvider) Get(realmID string) *RealmService {
	return NewRealmService(p.base, p, realmID)
}

func (p *RealmsServiceProvider) New(realmData *realm.Realm, key *jose.JsonWebKey) (*realm.Realm, error) {

	if key == nil {
		var err error
		key, err = crypto.NewKey()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create key for realm")
		}
	}

	if realmData.Name == "" {
		realmData.Name = fmt.Sprintf("%s.%s", crypto.Thumbprint(key), viper.GetString("proxy_domain"))
	}

	if realmData.ID == "" {
		realmData.ID = realmData.Name
	}

	re, err := regexp.Compile("^[0-9|a-z|A-Z||\\-\\.\\:]*$")
	if err != nil {
		return nil, errors.Wrap(err, "could not build regex matcher")
	}

	if !re.MatchString(realmData.Name) {
		return nil, errors.New("Bad realm name")
	}

	_, err = p.realms.Get(realmData.ID)
	if err == nil {
		return nil, errors.New("Realm already exists")
	}

	pk, err := crypto.NewPublicKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get public key")
	}

	if len(realmData.AdminRoles) < 1 {
		realmData.AdminRoles = []string{"admin@" + realmData.Name}
	}

	realmData.PublicKey = pk

	realmData.Descriptor = document.NewRealmDescriptor(realmData.ID, pk, fmt.Sprintf("%s/realm/v2/realms/%s/services", p.base, realmData.ID))
	realmData.Descriptor.Description = realmData.Description

	skey := keys.NewStoredKey(realmData.ID)
	if err := skey.Encrypt(key, p.kek); err != nil {
		return nil, errors.Wrap(err, "failed to encrypt private key for realm")
	}

	if err := p.sks.Save(skey); err != nil {
		return nil, errors.Wrap(err, "failed to save key for realm")
	}

	realmData.SignedDescriptor, err = p.signDescriptor(realmData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign descriptor")
	}

	if err := p.realms.Set(realmData); err != nil {
		return nil, errors.Wrap(err, "failed to save realm")
	}

	if err = p.publishEvent("ADDED", realmData.ID); err != nil {
		return nil, errors.Wrap(err, "failed to publish event for created realm")
	}

	return realmData, nil
}

func (p *RealmsServiceProvider) signDescriptor(realmData *realm.Realm) (string, error) {
	descBytes, err := json.Marshal(realmData.Descriptor)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal realm descriptor")
	}

	descSigned, err := p.signPayload(realmData.ID, descBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign descriptor")
	}

	return descSigned.FullSerialize(), nil
}

func (p *RealmsServiceProvider) getKey(realmID string) (*jose.JsonWebKey, error) {
	skey, err := p.sks.Get(realmID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key")
	}

	return skey.Decrypt(p.kek)
}

func (p *RealmsServiceProvider) signPayload(realmID string, payload []byte) (*jose.JsonWebSignature, error) {
	skey, err := p.sks.Get(realmID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key")
	}

	key, err := skey.Decrypt(p.kek)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt key")
	}

	signer, err := crypto.NewSigner(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create signer")
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign descriptor")
	}

	return jws, nil
}

func (p *RealmsServiceProvider) publishEvent(eventType, realmID string) error {
	event := &realm.RealmEvent{
		Type:  eventType,
		Realm: realmID,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err = p.pubsub.Publish(p.realmTopic, string(eventBytes)); err != nil {
		return err
	}

	return nil
}
