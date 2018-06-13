package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Brickchain/go-crypto.v2"
	logger "github.com/Brickchain/go-logger.v1"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitlab.brickchain.com/controllers/proxy/proxy"
	jose "gopkg.in/square/go-jose.v1"

	"github.com/Brickchain/go-document.v2"
)

type Proxy struct {
	base      string
	id        string
	url       string
	endpoint  string
	conn      *websocket.Conn
	writeLock *sync.Mutex
	regDone   chan error
	connected bool
	handler   http.Handler
	key       *jose.JsonWebKey
}

func NewProxy(endpoint string) (*Proxy, error) {
	p := &Proxy{
		endpoint:  endpoint,
		writeLock: &sync.Mutex{},
	}

	go p.subscribe()

	return p, nil
}

// type regData struct {
// 	Platform  string `json:"platform"`
// 	PushType  string `json:"pushType"`
// 	PushToken string `json:"pushToken"`
// 	AppID     string `json:"appID"`
// 	Token     string `json:"token,omitempty"`
// }

// type regResp struct {
// 	AppID        string `json:"appID"`
// 	WebsocketURL string `json:"websocketURL"`
// }

func (p *Proxy) connect() error {
	host := strings.Replace(strings.Replace(p.endpoint, "https://", "", 1), "http://", "", 1)
	schema := "ws"
	if strings.HasPrefix(p.endpoint, "https://") {
		schema = "wss"
	}

	u := url.URL{Scheme: schema, Host: host, Path: "/proxy/subscribe"}

	var err error
	p.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	p.connected = true

	return nil
}

func (p *Proxy) write(b []byte) error {
	p.writeLock.Lock()
	defer p.writeLock.Unlock()

	return p.conn.WriteMessage(websocket.TextMessage, b)
}

func (p *Proxy) Register(key *jose.JsonWebKey) error {

	p.id = crypto.Thumbprint(key)
	p.key = key

	return nil

}

func (p *Proxy) register() error {

	for {
		if p.connected {
			break
		}

		time.Sleep(time.Millisecond * 10)
	}

	p.regDone = make(chan error)

	mandateToken := document.NewMandateToken([]string{}, p.endpoint, 60)

	b, _ := json.Marshal(mandateToken)

	signer, err := crypto.NewSigner(p.key)
	if err != nil {
		return err
	}

	jws, err := signer.Sign(b)
	if err != nil {
		return err
	}

	jwsCompact, _ := jws.CompactSerialize()

	regReq := proxy.NewRegistrationRequest(jwsCompact)
	regReqBytes, _ := json.Marshal(regReq)

	if err := p.write(regReqBytes); err != nil {
		return err
	}

	if err := <-p.regDone; err != nil {
		return err
	}

	p.base = fmt.Sprintf("%s.%s", p.id, viper.GetString("proxy_domain"))

	return nil
}

func (p *Proxy) SetHandler(handler http.Handler) {
	p.handler = handler
}

func (p *Proxy) subscribe() error {
	disconnect := func() {
		if p.connected {
			p.conn.Close()
			p.connected = false
		}
	}

	for {
		if !p.connected {
			logger.Debug("Connecting to proxy...")
			if err := p.connect(); err != nil {
				logger.Error(errors.Wrap(err, "failed to connect to proxy"))
				disconnect()
				time.Sleep(time.Second * 10)
				continue
			}
			logger.Debug("Connected!")

			if p.key != nil {
				go func() {
					logger.Debug("Registering to proxy...")
					if err := p.register(); err != nil {
						logger.Error(errors.Wrap(err, "failed to register to proxy"))
						disconnect()
					} else {
						logger.Debug("Registered!")
					}
				}()
			}
		}

		_, body, err := p.conn.ReadMessage()
		if err != nil {
			logger.Error(errors.Wrap(err, "failed to read message"))
			disconnect()
			continue
		}
		// logger.Infof("recv: %s", body)

		docType, err := document.GetType(body)
		if err != nil {
			logger.Error(errors.Wrap(err, "failed to get document type"))
		}

		switch docType {
		case proxy.SchemaBase + "/registration-response.json":
			r := &proxy.RegistrationResponse{}
			if err := json.Unmarshal(body, &r); err != nil {
				logger.Error(errors.Wrap(err, "failed to unmarshal registration-response"))
			}

			if r.KeyID != p.id {
				p.regDone <- errors.New("Wrong KeyID in registration")
			} else {
				close(p.regDone)
			}

		case proxy.SchemaBase + "/http-request.json":

			if p.handler == nil {
				logger.Error("No handler set, can't process http-request")
				continue
			}

			req := &proxy.HttpRequest{}
			if err := json.Unmarshal(body, &req); err != nil {
				logger.Error(errors.Wrap(err, "failed to unmarshal http-request"))
				continue
			}

			if req != nil {
				r := &http.Request{
					Method: req.Method,
					URL: &url.URL{
						Host: p.base,
						Path: req.URL,
					},
					RequestURI: req.URL,
					Header:     make(http.Header),
					Host:       req.Headers["Host"],
				}

				if req.Headers["X-Forwarded-Host"] != "" {
					r.Host = req.Headers["X-Forwarded-Host"]
				}

				for k, v := range req.Headers {
					r.Header.Set(k, v)
				}

				if req.Body != "" {
					r.Body = nopCloser{bytes.NewBufferString(req.Body)}
				}

				w := httptest.NewRecorder()

				p.handler.ServeHTTP(w, r)

				res := proxy.NewHttpResponse(req.ID, w.Result().StatusCode)
				res.ContentType = w.Result().Header.Get("Content-Type")

				body, _ := ioutil.ReadAll(w.Result().Body)
				res.Body = string(body)

				res.Headers = make(map[string]string)
				for k, v := range w.Result().Header {
					res.Headers[k] = v[0]
				}

				b, _ := json.Marshal(res)
				if err := p.write(b); err != nil {
					logger.Error(errors.Wrap(err, "failed to send http-response"))
					disconnect()
					continue
				}
			}
		}
	}
}

// type HttpRequest struct {
// 	URL     string            `json:"url"`
// 	Headers map[string]string `json:"headers"`
// 	Method  string            `json:"method"`
// 	Body    string            `json:"body"`
// }

// type HttpResponse struct {
// 	ID          string            `json:"id"`
// 	Headers     map[string]string `json:"headers"`
// 	ContentType string            `json:"contentType"`
// 	Status      int               `json:"status"`
// 	Body        string            `json:"body"`
// }

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}
