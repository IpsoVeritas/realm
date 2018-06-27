package messaging

import (
	"net/url"

	"regexp"

	"errors"

	"bytes"
	"text/template"

	"github.com/spf13/viper"
	"github.com/subosito/twilio"
)

var validE164 *regexp.Regexp

type TwilioTransport struct {
	From   string
	twilio *twilio.Client
}

func NewTwilioTransport(config *viper.Viper) (t *TwilioTransport) {
	validE164 = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	accountSid := config.GetString("twilio_account_sid")
	authToken := config.GetString("twilio_authtoken")
	tc := twilio.NewClient(accountSid, authToken, nil)
	t = &TwilioTransport{config.GetString("twilio_from"), tc}
	return
}

func (t *TwilioTransport) Validate(m Message) (err error) {
	if m.Templates.Text == "" {
		return errors.New("Text is empty")
	}
	u, err := url.Parse(m.Recipient)
	if err == nil && !validE164.MatchString(u.Opaque) {
		err = errors.New(u.Opaque + " does not follow the E.164 format")
	}

	if len(m.Attachments) > 0 {
		err = errors.New("Transport does not support attachments")
	}
	return
}

func (t *TwilioTransport) Send(m Message) (id string, err error) {

	var textBuffer bytes.Buffer
	textTemplate := template.Must(template.New("text").Parse(m.Templates.Text))
	textTemplate.Execute(&textBuffer, m.Data)

	u, _ := url.Parse(m.Recipient)

	twilioMessage, _, err := t.twilio.Messages.SendSMS(t.From, u.Opaque, textBuffer.String())

	if twilioMessage != nil {
		id = twilioMessage.Sid
	}

	return

}
