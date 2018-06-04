package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/gorilla/websocket"
	jose "gopkg.in/square/go-jose.v1"

	"github.com/Brickchain/go-document.v2"
)

type Proxy struct {
	base     string
	key      *jose.JsonWebKey
	url      string
	endpoint string
	token    string
}

func NewProxy(key *jose.JsonWebKey, endpoint, token string) (*Proxy, error) {
	p := &Proxy{
		key:      key,
		endpoint: endpoint,
		token:    token,
	}

	return p, p.register()
}

type regData struct {
	Platform  string `json:"platform"`
	PushType  string `json:"pushType"`
	PushToken string `json:"pushToken"`
	AppID     string `json:"appID"`
	Token     string `json:"token,omitempty"`
}

type regResp struct {
	AppID        string `json:"appID"`
	WebsocketURL string `json:"websocketURL"`
}

func (p *Proxy) register() error {
	reg := regData{
		Platform: "web",
		PushType: "backend",
		AppID:    "com.brickchain.integrity",
		Token:    p.token,
	}

	b, _ := json.Marshal(reg)

	signer, err := crypto.NewSigner(p.key)
	if err != nil {
		return err
	}

	jws, err := signer.Sign(b)
	if err != nil {
		return err
	}

	r, err := http.Post(p.endpoint, "application/json", bytes.NewReader([]byte(jws.FullSerialize())))
	if err != nil {
		return err
	}

	rb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	resp := &regResp{}
	if err := json.Unmarshal(rb, &resp); err != nil {
		return err
	}

	p.base = resp.AppID
	p.url = resp.WebsocketURL

	return nil
}

func (p *Proxy) Subscribe(handler http.Handler) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial(p.url, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}
		// logger.Infof("recv: %s", message)

		var pm *document.PushMessage
		if err := json.Unmarshal(message, &pm); err != nil {
			continue
		}

		var req *HttpRequest
		if err := json.Unmarshal([]byte(pm.Data), &req); err != nil {
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

			handler.ServeHTTP(w, r)

			res := &HttpResponse{
				ID:          pm.ID,
				Status:      w.Result().StatusCode,
				ContentType: w.Result().Header.Get("Content-Type"),
			}

			body, _ := ioutil.ReadAll(w.Result().Body)
			res.Body = string(body)

			res.Headers = make(map[string]string)
			for k, v := range w.Result().Header {
				res.Headers[k] = v[0]
			}

			b, _ := json.Marshal(res)
			if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
				return err
			}
		}
	}
}

type HttpRequest struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	Body    string            `json:"body"`
}

type HttpResponse struct {
	ID          string            `json:"id"`
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"contentType"`
	Status      int               `json:"status"`
	Body        string            `json:"body"`
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}
