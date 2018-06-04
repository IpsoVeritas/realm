package document

const PushMessageType = "push-message"

type PushMessage struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	URI     string `json:"uri,omitempty"`
	Data    string `json:"data,omitempty"`
}

func NewPushMessage(title, message string) *PushMessage {
	return &PushMessage{
		Title:   title,
		Message: message,
	}
}
