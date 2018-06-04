package realm

import (
	messaging "gitlab.brickchain.com/libs/go-messaging.v1"
)

type EmailStatus struct {
	MessageID   string   `json:"messageID,omitempty"`
	Sent        bool     `json:"sent"`
	Subject     string   `json:"subject,omitempty"`
	Rendered    string   `json:"rendered,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type EmailProvider interface {
	Validate(messaging.Message) error
	Send(messaging.Message) (*EmailStatus, error)
}
