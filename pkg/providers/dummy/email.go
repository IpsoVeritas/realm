package dummy

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"

	logger "github.com/Brickchain/go-logger.v1"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	messaging "gitlab.brickchain.com/libs/go-messaging.v1"
)

type DummyEmailProvider struct{}

func NewDummyEmailProvider() (realm.EmailProvider, error) {
	return &DummyEmailProvider{}, nil
}

func (p *DummyEmailProvider) Send(msg messaging.Message) (*realm.EmailStatus, error) {
	status := p.render(msg)
	status.Sent = false

	return status, nil
}

func (p *DummyEmailProvider) Validate(msg messaging.Message) error {
	if msg.Templates.Subject == "" {
		return errors.New("Subject is empty")
	}
	if msg.Templates.Text == "" {
		return errors.New("Text is empty")
	}
	u, err := url.Parse(msg.Recipient)
	if err == nil {
		if _, err := mail.ParseAddress(u.Opaque); err != nil {
			return err
		}
	}

	return nil
}

func (p *DummyEmailProvider) render(msg messaging.Message) *realm.EmailStatus {

	status := &realm.EmailStatus{}

	var subjectBuffer, textBuffer bytes.Buffer

	subjectTemplate := template.Must(template.New("subject").Parse(msg.Templates.Subject))
	subjectTemplate.Execute(&subjectBuffer, msg.Data)
	logger.Debugf("Subject: %s", subjectBuffer.String())
	status.Subject = subjectBuffer.String()

	textTemplate := template.Must(template.New("text").Parse(msg.Templates.Text))
	textTemplate.Execute(&textBuffer, msg.Data)
	logger.Debugf("Text: %s", textBuffer.String())
	status.Rendered = textBuffer.String()

	status.Attachments = make(map[string]string)
	for _, f := range msg.Attachments {
		b, err := ioutil.ReadFile(f)
		if err == nil {
			key := fmt.Sprintf("cid:%s", filepath.Base(f))
			base64 := base64.StdEncoding.EncodeToString(b)
			imageType := strings.Replace(filepath.Ext(f), ".", "/", 1)
			status.Attachments[key] = fmt.Sprintf("data:image%s;base64,%s", imageType, base64)
		} else {
			logger.Warningf("Failed to read attachment %s", f)
		}
	}

	if msg.Templates.HTML != "" {
		var htmlBuffer bytes.Buffer
		htmlTemplate := template.Must(template.New("html").Parse(msg.Templates.HTML))
		htmlTemplate.Execute(&htmlBuffer, msg.Data)
		logger.Debugf("HTML: %s", htmlBuffer.String())
		status.Rendered = htmlBuffer.String()
	}

	return status
}
