package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
	jose "gopkg.in/square/go-jose.v1"
)

type RealmsController struct {
	base            string
	contextProvider *services.RealmsServiceProvider
	keyset          *jose.JsonWebKeySet
}

func NewRealmsController(
	base string,
	contextProvider *services.RealmsServiceProvider,
	keyset *jose.JsonWebKeySet,
) *RealmsController {

	r := &RealmsController{
		base:            base,
		contextProvider: contextProvider,
		keyset:          keyset,
	}

	return r
}

func (c *RealmsController) ListRealms(req httphandler.AuthenticatedRequest) httphandler.Response {

	if c.contextProvider.HasMandateForBootstrapRealm(req.Mandates()) {

		realms, err := c.contextProvider.ListRealms()
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to list realms"))
		}

		return httphandler.NewJsonResponse(http.StatusOK, realms)

	} else {

		realms := make([]*realm.Realm, 0)

		for _, m := range req.Mandates() {
			context := c.contextProvider.Get(m.Mandate.Realm)

			realm, err := context.Realm()
			if err == nil {
				realms = append(realms, realm)
			}
		}

		return httphandler.NewJsonResponse(http.StatusOK, realms)

	}

}

func (c *RealmsController) GetRealm(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("api.v2.realms.GetRealm.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Need to specify realm"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	realm, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get realm"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, realm)
}

func (c *RealmsController) PostRealm(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("api.v2.realms.PostRealm.total")
	defer total.Stop()

	if !c.contextProvider.HasMandateForBootstrapRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No access to create realms"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body"))
	}

	realm := &realm.Realm{}
	if err := json.Unmarshal(body, &realm); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to unmarshal realm json"))
	}

	re, err := regexp.Compile("^[0-9|a-z|\\-\\.]*$")
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not build regex matcher"))
	}

	if !re.MatchString(realm.ID) {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Bad realm ID"))
	}

	_, err = c.contextProvider.Get(realm.ID).Realm()
	if err == nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Realm already exists"))
	}

	createdRealm, err := c.contextProvider.New(realm, nil)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to create realm"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, createdRealm)
}

func (c *RealmsController) UpdateRealm(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("api.v2.realms.UpdateRealm.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body"))
	}

	realm := &realm.Realm{}
	if err := json.Unmarshal(body, &realm); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to unmarshal realm json"))
	}

	realm.Descriptor.Label = realm.Label

	if err := context.Set(realm); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to save realm"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, realm)
}

func (c *RealmsController) DeleteRealm(req httphandler.AuthenticatedRequest) httphandler.Response {

	total := stats.StartTimer("api.realms.DeleteRealm.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	if err := context.Delete(); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete realm"))
	}

	return httphandler.NewEmptyResponse(http.StatusNoContent)
}

// // ===============================================================
// // this method is publicly accessible.
// //
func (c *RealmsController) JoinRealm(req httphandler.Request) httphandler.Response {

	total := stats.StartTimer("api.realms.JoinRealm.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to read request body"))
	}

	jws, err := crypto.UnmarshalSignature(body)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to parse JWS"))
	}

	multipart, err := context.Join(jws)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to join realm"))
	}

	return httphandler.NewJsonResponse(http.StatusOK, multipart)
}

func (c *RealmsController) JoinRealmCallback(req httphandler.Request) httphandler.Response {

	total := stats.StartTimer("api.realms.JoinRealmCallback.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	realmData, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not find realm"))
	}

	if realmData.GuestRole == "" {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No public invite for realm"))
	}

	guestRole, err := context.Roles().ByName(realmData.GuestRole)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not get guest role"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to read request body"))
	}

	jws, err := crypto.UnmarshalSignature(body)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to parse JWS"))
	}

	if len(jws.Signatures) < 1 {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No signatures on JWS"))
	}

	var userKey *jose.JsonWebKey
	var scopeResponse document.Multipart
	var certificate *document.Certificate

	scopeDataBytes, err := jws.Verify(jws.Signatures[0].Header.JsonWebKey)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify JWS signature"))
	}

	if err := json.Unmarshal(scopeDataBytes, &scopeResponse); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal multipart in JWS"))
	}

	if scopeResponse.Certificate != "" {
		certificate, err = crypto.VerifyCertificate(scopeResponse.Certificate, guestRole.KeyLevel)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify certificate chain"))
		}
	}

	if certificate != nil {
		userKey = certificate.Issuer
	} else {
		userKey = jws.Signatures[0].Header.JsonWebKey
	}

	var contract *document.Contract
	userData := make(map[string]interface{})
	for _, part := range scopeResponse.Parts {
		if part.Name == "contract" {
			if err := json.Unmarshal([]byte(part.Document), &contract); err != nil {
				req.Log().Error(err)
			}
		} else if part.Encoding == "application/json+jws" {
			partJWS, err := crypto.UnmarshalSignature([]byte(part.Document))
			if err == nil && len(partJWS.Signatures) > 0 {
				ok := validSig(c.keyset, partJWS)
				// TODO: more info on matching keys
				if ok {
					b, err := partJWS.Verify(partJWS.Signatures[0].Header.JsonWebKey)
					if err == nil {
						var fact document.Fact
						err := json.Unmarshal(b, &fact)
						if err == nil {
							for k, v := range fact.Data {
								userData[k] = v
							}
						}
					}
				} else {
					req.Log().Warn("No matching keys for TODO")
				}
			}
		}
	}

	if contract == nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("User did not accept the contract"))
	}

	// for key, value := range ticket.Facts {
	// 	userValue, ok := userData[key]
	// 	if ok {
	// 		if userValue != value {
	// 			errStr := fmt.Sprintf("Fact %s does not match value provided by user", key)
	// 			logger.Warn(errStr)
	// 			http.Error(w, errStr, http.StatusBadRequest)
	// 			return
	// 		}
	// 	} else {
	// 		errStr := fmt.Sprintf("Fact %s not provided by user", key)
	// 		logger.Warn(errStr)
	// 		http.Error(w, errStr, http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// TODO: get users root key when we have key lists
	mandate := document.NewMandate(guestRole.Name)
	mandate.RoleName = guestRole.Description
	mandate.Recipient = userKey
	mandate.Sender = realmID
	mandate.Realm = realmID

	// if userName, ok := userData["name"]; ok {
	// 	m.RecipientName = userName.(string)
	// }
	// if m.RecipientName == "" {
	// 	if userEmail, ok := userData["email"]; ok {
	// 		m.RecipientName = userEmail.(string)
	// 	}
	// }
	// if m.RecipientName == "" {
	// 	if userPhone, ok := userData["phone"]; ok {
	// 		m.RecipientName = userPhone.(string)
	// 	}
	// }
	// if m.RecipientName == "" {
	// 	m.RecipientName = userID
	// }

	issued, err := context.Mandates().Issue(mandate, userKey.KeyID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to issue mandate"))
	}

	part := document.Part{
		Encoding: "application/json+jws",
		Name:     "mandate",
		Document: issued.Signed,
	}
	multipart := document.NewMultipart()
	multipart.Append(part)

	multipartBytes, err := json.Marshal(multipart)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to marshal response"))
	}

	return httphandler.NewJsonResponse(http.StatusCreated, multipartBytes)
}

func (c *RealmsController) IconHandler(req httphandler.AuthenticatedRequest) httphandler.Response {

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	realm, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not find realm"))
	}

	file, handler, err := req.OriginalRequest().FormFile("file")
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get file from request"))
	}
	defer file.Close()

	if handler.Filename == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No filename"))
	}

	ext := strings.Split(handler.Filename, ".")
	if len(ext) < 2 {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("File name provided does not contain a file type"))
	}
	filetype := ext[len(ext)-1]

	name, err := context.Files().Write(fmt.Sprintf("icon.%s", filetype), file)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to write file to storage"))
	}

	realm.Descriptor.Icon = name

	if err := context.Set(realm); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to save changes"))
	}

	return httphandler.NewEmptyResponse(http.StatusCreated)
}

func (c *RealmsController) BannerHandler(req httphandler.AuthenticatedRequest) httphandler.Response {

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	if !context.HasMandateForRealm(req.Mandates()) {
		return httphandler.NewErrorResponse(http.StatusForbidden, errors.New("No mandate for realm"))
	}

	realm, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "could not find realm"))
	}

	file, handler, err := req.OriginalRequest().FormFile("file")
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get file from request"))
	}
	defer file.Close()

	if handler.Filename == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No filename"))
	}

	ext := strings.Split(handler.Filename, ".")
	if len(ext) < 2 {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("File name provided does not contain a file type"))
	}
	filetype := ext[len(ext)-1]

	name, err := context.Files().Write(fmt.Sprintf("banner.%s", filetype), file)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to write file to storage"))
	}

	realm.Descriptor.Banner = name

	if err := context.Set(realm); err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to save realm"))
	}

	return httphandler.NewEmptyResponse(http.StatusCreated)
}

func (c *RealmsController) Bootstrap(req httphandler.Request) httphandler.Response {
	req.Log().Debug("bootstrap!")

	total := stats.StartTimer("api.access.Bootstrap.total")
	defer total.Stop()

	password, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "could not read password"))
	}

	if string(password) == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Sent password is empty"))
	}

	ticket, err := c.contextProvider.Bootstrap(string(password))
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to bootstrap"))
	}

	url := document.URLResponse{
		URL: fmt.Sprintf("%s/realm/v2/realms/%s/tickets/%s/issue", c.base, ticket.Realm, ticket.ID),
	}

	return httphandler.NewJsonResponse(http.StatusCreated, url)
}
