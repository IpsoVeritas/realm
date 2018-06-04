package document

import "time"

type DocumentCallback struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	LinkID    string    `json:"linkId,omitempty"`
	LinkType  string    `json:"linkType,omitempty"`
}

func NewDocumentCallback() DocumentCallback {
	return DocumentCallback{
		Timestamp: time.Now(),
	}
}
