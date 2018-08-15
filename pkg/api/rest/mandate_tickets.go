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

	var certificate *document.Certificate

	if mp.Certificate != "" {
		certificate, err = crypto.VerifyCertificate(mp.Certificate, 10)
		if err != nil {
			return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "failed to verify certificate"))
		}
	}

	if certificate != nil {
		userKey = certificate.Issuer
	} else {
		userKey = jws.Signatures[0].Header.JsonWebKey
	}

	name := userKey.KeyID
	for _, part := range mp.Parts {

		if part.Name == "name" {

			var fact document.Fact
			err := json.Unmarshal([]byte(part.Document), &fact)
			if err != nil {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal fact"))
			}

			fmt.Printf("\n\nFACT => %+v\n\n", fact)

			if len(fact.Signatures) < 1 {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No signatures on fact"))
			}

			for _, sig := range fact.Signatures {

				sigJWS, err := crypto.UnmarshalSignature([]byte(sig))
				if err != nil {
					return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal signature"))
				}

				b, err := sigJWS.Verify(sigJWS.Signatures[0].Header.JsonWebKey)
				if err != nil {
					return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "Failed to verify signature"))
				}

				var factSignature document.Fact
				err = json.Unmarshal([]byte(b), &factSignature)
				if err != nil {
					return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal fact signature"))
				}

			}

			var ok bool
			name, ok = fact.Data["name"].(string)
			if !ok {
				return httphandler.NewErrorResponse(http.StatusBadRequest, errors.Wrap(err, "Invalid type"))
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
