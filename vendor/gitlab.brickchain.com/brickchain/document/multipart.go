package document

import (
	"sync"
	"time"
)

const MultipartType = "multipart"

type Multipart struct {
	Base
	Parts []Part `json:"parts"`
}

type Part struct {
	Encoding string `json:"encoding,omitempty"`
	Name     string `json:"name,omitempty"`
	Document string `json:"document,omitempty"`
	URI      string `json:"uri,omitempty"`
}

func NewMultipart() *Multipart {
	m := &Multipart{
		Base: Base{
			Context:   Context,
			Type:      MultipartType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		Parts: make([]Part, 0),
	}

	return m
}

func (m *Multipart) Append(part Part) {
	m.Parts = append(m.Parts, part)
}

func (m *Multipart) AppendDoc(doc BaseInterface) {
	d, e := Serialize(doc)
	p := Part{
		Encoding: e,
		Document: d,
	}
	m.Append(p)
}
