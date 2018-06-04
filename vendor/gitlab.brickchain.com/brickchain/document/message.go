package document

import "time"

const MessageType = "message"

type Message struct {
	Base
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewMessage(title, message string) *Message {
	return &Message{
		Base: Base{
			Context:   Context,
			Type:      MessageType,
			Timestamp: time.Now(),
		},
		Title:   title,
		Message: message,
	}
}
