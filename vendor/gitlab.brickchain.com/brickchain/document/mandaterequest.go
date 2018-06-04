package document

import (
	"sync"
	"time"
)

const MandateRequestType = "mandate-request"

type MandateRequest struct {
	Base
	ReplyTo   []string `json:"replyTo"`
	Roles     []string `json:"scopes"`
	EncryptTo []string `json:"encryptTo,omitempty"`
}

func NewMandateRequest() *MandateRequest {
	m := &MandateRequest{
		Base: Base{
			Context:   Context,
			Type:      MandateRequestType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		ReplyTo:   []string{},
		Roles:     []string{},
		EncryptTo: []string{},
	}
	return m
}
