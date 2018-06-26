package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	crypto "github.com/Brickchain/go-crypto.v2"
	document "github.com/Brickchain/go-document.v2"
	"github.com/pkg/errors"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	messaging "gitlab.brickchain.com/libs/go-messaging.v1"
	jose "gopkg.in/square/go-jose.v1"
)

type InviteService struct {
	base    string
	p       realm.InviteProvider
	realmID string
	realm   *RealmService
	email   realm.EmailProvider
	assets  realm.AssetProvider
}

func (i *InviteService) List() ([]*realm.Invite, error) {
	return i.p.List(i.realmID)
}

func (i *InviteService) Get(id string) (*realm.Invite, error) {
	return i.p.Get(i.realmID, id)
}

func (i *InviteService) Set(invite *realm.Invite) error {
	invite.Realm = i.realmID
	return i.p.Set(i.realmID, invite)
}

func (i *InviteService) Delete(id string) error {
	return i.p.Delete(i.realmID, id)
}

func (i *InviteService) ListForRole(role string) ([]*realm.Invite, error) {
	return i.p.ListForRole(i.realmID, role)
}

func (i *InviteService) Send(invite *realm.Invite) (*realm.EmailStatus, error) {
	role, err := i.realm.Roles().ByName(invite.Role)
	if err != nil {
		return nil, err
	}

	templateDir := "invite_email"
	templateFile := func(name string) string {
		return fmt.Sprintf("%s/%s", templateDir, name)
	}
	attachmentDir := templateFile("attachments")
	attachmentFile := func(name string) string {
		return fmt.Sprintf("%s/%s", attachmentDir, name)
	}

	u := fmt.Sprintf("%s/realm/v2/realms/%s/invites/id/%s/fetch", i.base, i.realmID, invite.ID)
	encoded := url.QueryEscape(u)
	link := fmt.Sprintf("https://app.plusintegrity.com?data=%s", encoded)

	templateSubject := `Invitation to join {{ .realm }} as {{ .roleName }}`

	templateHTML, err := i.assets.Read(templateFile("template.html"))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't read HTML email template")
	}

	templateText, err := i.assets.Read(templateFile("template.txt"))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't read text email template")
	}

	realmData, err := i.realm.Realm()
	if err != nil {
		return nil, err
	}

	message := messaging.Message{
		Recipient: invite.MessageURI,
		Templates: messaging.Templates{
			Subject: templateSubject,
			Text:    string(templateText),
			HTML:    string(templateHTML),
		},
		Data: map[string]interface{}{
			"role":     invite.Role,
			"roleName": role.Description,
			"realm":    realmData.Description,
			"url":      u,
			"text":     invite.Text,
			"link":     link,
		},
	}

	attachments, err := i.assets.List(attachmentDir)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't list attachments directory")
	}

	cleanup := make([]string, 0)
	defer func() {
		for _, f := range cleanup {
			os.RemoveAll(f)
		}
	}()

	if len(attachments) > 0 {
		message.Attachments = make([]string, 0)
		for _, attachment := range attachments {
			filename, err := i.assets.CopyToTempFile(attachmentFile(attachment))
			// logger.Errorf("Got file %s", filename)
			if err != nil {
				// logger.Error(err)
				return nil, errors.Wrapf(err, "Could not get attachment %s", attachment)
			}
			cleanup = append(cleanup, filepath.Dir(filename))
			message.Attachments = append(message.Attachments, filename)
		}
	}

	if err := i.email.Validate(message); err != nil {
		return nil, errors.Wrap(err, "failed to validate message")
	}

	return i.email.Send(message)
}

func (i *InviteService) Fetch(inviteID string) (*jose.JsonWebSignature, error) {
	invite, err := i.Get(inviteID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invite")
	}

	scopeRequest := document.NewScopeRequest(invite.KeyLevel)
	scopeRequest.ReplyTo = []string{
		fmt.Sprintf("%s/realm/v2/realms/%s/invites/id/%s/callback", i.base, i.realmID, inviteID),
	}
	scopeRequest.KeyLevel = invite.KeyLevel

	scopeRequest.Contract = document.NewContract()
	scopeRequest.Contract.Text = fmt.Sprintf("Receive mandate for %s", invite.Role)

	scopeReqBytes, err := json.Marshal(scopeRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal scope-request")
	}

	jws, err := i.realm.Sign(scopeReqBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign scope-request")
	}

	return jws, nil
}

func (i *InviteService) Callback(inviteID string, jws *jose.JsonWebSignature) (*document.Multipart, error) {

	invite, err := i.Get(inviteID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invite")
	}

	var userKey *jose.JsonWebKey
	var scopeResponse document.Multipart
	var certificate *document.Certificate

	if len(jws.Signatures) < 1 {
		return nil, errors.Wrap(err, "no key in signature")
	}

	scopeDataBytes, err := jws.Verify(jws.Signatures[0].Header.JsonWebKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify signature")
	}

	if err := json.Unmarshal(scopeDataBytes, &scopeResponse); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal scope-request")
	}

	if scopeResponse.Certificate != "" {
		certificate, err = crypto.VerifyCertificate(scopeResponse.Certificate, invite.KeyLevel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to verify certificate")
		}
	}

	if certificate != nil {
		userKey = certificate.Issuer
	} else {
		userKey = jws.Signatures[0].Header.JsonWebKey
	}

	role, err := i.realm.Roles().ByName(invite.Role)
	if err != nil {
		return nil, errors.Wrap(err, "could not get role")
	}

	mandate := document.NewMandate(invite.Role)
	mandate.Realm = i.realmID
	mandate.RoleName = role.Description
	mandate.ValidFrom = invite.ValidFrom
	mandate.ValidUntil = invite.ValidUntil
	mandate.Recipient = userKey
	mandate.Sender = invite.Sender

	issued, err := i.realm.Mandates().Issue(mandate, invite.Name)
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

	if err := i.Delete(invite.ID); err != nil {
		return nil, errors.Wrap(err, "failed to delete invite")
	}

	return multipart, nil
}
