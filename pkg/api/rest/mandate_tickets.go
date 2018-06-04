package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/pkg/errors"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
	jose "gopkg.in/square/go-jose.v1"
)

type MandateTicketController struct {
	base            string
	contextProvider *services.RealmsServiceProvider
	keyset          *jose.JsonWebKeySet
}

func NewMandateTicketController(
	base string,
	contextProvider *services.RealmsServiceProvider,
) *MandateTicketController {

	r := &MandateTicketController{
		base:            base,
		contextProvider: contextProvider,
	}

	return r
}

func (c *MandateTicketController) IssueMandate(req httphandler.Request) httphandler.Response {

	total := stats.StartTimer("api.mandate_tickets.IssueMandate.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	ticketID := req.Params().ByName("ticketID")
	if ticketID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No ticket ID specified"))
	}

	context := c.contextProvider.Get(realmID)

	ticket, err := context.MandateTickets().Get(ticketID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusNotFound, errors.Wrap(err, "ticket not found"))
	}

	ticket.ScopeRequest.ReplyTo = []string{fmt.Sprintf("%s/realm/v2/realms/%s/tickets/%s/callback", c.base, ticket.Realm, ticket.ID)}

	scopeReqBytes, err := json.Marshal(ticket.ScopeRequest)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to marshal response"))
	}

	scopeReqJWS, err := context.Sign(scopeReqBytes)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to sign response"))
	}

	return httphandler.NewStandardResponse(http.StatusOK, "application/json", scopeReqJWS.FullSerialize())
}

func (c *MandateTicketController) IssueMandateCallback(req httphandler.Request) httphandler.Response {

	total := stats.StartTimer("api.mandate_tickets.IssueMandateCallback.total")
	defer total.Stop()

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	ticketID := req.Params().ByName("ticketID")
	if ticketID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No ticket ID specified"))
	}

	context := c.contextProvider.Get(realmID)

	ticket, err := context.MandateTickets().Get(ticketID)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusNotFound, errors.Wrap(err, "ticket not found"))
	}

	body, err := req.Body()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to read request body"))
	}

	jws, err := crypto.UnmarshalSignature(body)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal JWS"))
	}

	if len(jws.Signatures) < 1 {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No signatures on JWS"))
	}

	payload, err := jws.Verify(jws.Signatures[0].Header.JsonWebKey)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify JWS signature"))
	}

	var userKey *jose.JsonWebKey
	var mp document.Multipart

	if err := json.Unmarshal(payload, &mp); err != nil {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal multipart"))
	}

	if mp.Certificate != "" {
		verified, _, subject, err := crypto.VerifyDocumentWithCertificateChain(mp, ticket.ScopeRequest.KeyLevel)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify certificate chain"))
		}

		if !verified {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Could not verify certificate chain"))
		}

		userKey = subject
	} else {
		userKey = jws.Signatures[0].Header.JsonWebKey
	}

	name := userKey.KeyID
	for _, part := range mp.Parts {
		if part.Name == "name" {
			partJWS, err := crypto.UnmarshalSignature([]byte(part.Document))
			if len(partJWS.Signatures) < 1 {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No signatures on fact"))
			}

			b, err := partJWS.Verify(partJWS.Signatures[0].Header.JsonWebKey)
			if err != nil {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify fact signature"))
			}

			var fact document.Fact
			if err := json.Unmarshal(b, &fact); err != nil {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal fact"))
			}

			if fact.Data == nil {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No data in fact"))
			}

			var ok bool
			name, ok = fact.Data["name"].(string)
			if !ok {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("Fact does not contain name"))
			}
		}
	}

	mandate := ticket.Mandate
	mandate.Recipient = userKey

	issued, err := context.Mandates().Issue(mandate, name)
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to issue mandate"))
	}

	respMp := document.NewMultipart()
	respMp.Append(document.Part{
		Name:     "mandate",
		Encoding: "text/plain+jws",
		Document: issued.Signed,
	})

	if !ticket.Static {
		if err := context.MandateTickets().Delete(ticket.ID); err != nil {
			return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to delete ticket after issuing mandate"))
		}
	}

	return httphandler.NewJsonResponse(http.StatusOK, respMp)
}
