package realm

import (
	messaging "github.com/Brickchain/realm/pkg/providers/messaging"
)

type EmailStatus struct {
	MessageID   string            `json:"messageID,omitempty"`
	Sent        bool              `json:"sent"`
	Subject     string            `json:"subject,omitempty"`
	Rendered    string            `json:"rendered,omitempty"`
	Attachments map[string]string `json:"attachments,omitempty"`
}

type EmailProvider interface {
	Validate(messaging.Message) error
	Send(messaging.Message) (*EmailStatus, error)
}
