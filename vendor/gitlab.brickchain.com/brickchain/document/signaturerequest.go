package document

import (
	"sync"
	"time"
)

const SignatureRequestType = "signature-request"

type SignatureRequest struct {
	Base
	ReplyTo  []string `json:"replyTo"`
	Document string   `json:"document"`
	KeyLevel int      `json:"keyLevel,omitempty"`
}

func NewSignatureRequest(keyLevel int) *SignatureRequest {
	s := &SignatureRequest{
		Base: Base{
			Context:   Context,
			Type:      SignatureRequestType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		ReplyTo:  []string{},
		KeyLevel: keyLevel,
	}
	return s
}

const SignatureResponseType = "signature-response"

type SignatureResponse struct {
	Base
	Document string `json:"document"`
}

func NewSignatureResponse() *SignatureResponse {
	s := &SignatureResponse{
		Base: Base{
			Context:   Context,
			Type:      SignatureResponseType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
	}
	return s
}
