package messaging

import (
	"errors"

	"net/url"

	"sync"
)

type Templates struct {
	Subject string `json:"subject"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
}

type Message struct {
	Recipient   string                 `json:"recipient"`
	Templates   Templates              `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Attachments []string               `json:"attachments,omitempty"`
}

type TransportInterface interface {
	Validate(Message) (err error)
	Send(Message) (id string, err error)
}

var (
	transports map[string]TransportInterface
	mu         *sync.Mutex
)

func init() {
	mu = &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	transports = make(map[string]TransportInterface)
}

func AddTransport(scheme string, t TransportInterface) {
	mu.Lock()
	defer mu.Unlock()
	transports[scheme] = t
}

func LookupTransport(m Message) (t TransportInterface, err error) {
	u, err := url.Parse(m.Recipient)
	if err == nil {
		var prs bool
		t, prs = transports[u.Scheme]
		if !prs {
			err = errors.New("transport for '" + u.Scheme + "' not found")
		}
	}
	return
}
