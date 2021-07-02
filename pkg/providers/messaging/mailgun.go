package messaging

import (
	"errors"
	"net/mail"

	"net/url"

	"bytes"
	"text/template"

	"github.com/mailgun/mailgun-go"
	"github.com/spf13/viper"
)

type MailgunTransport struct {
	From     string
	mailgun  mailgun.Mailgun
	testMode bool
}

func NewMailgunTransport(config *viper.Viper) (t *MailgunTransport) {
	domain := config.GetString("mailgun_domain")
	apiKey := config.GetString("mailgun_api_key")
	publicAPIKey := config.GetString("mailgun_public_api_key")
	mg := mailgun.NewMailgun(domain, apiKey, publicAPIKey)
	t = &MailgunTransport{config.GetString("mailgun_from"), mg, config.GetBool("mailgun_testmode")}
	return
}

func (t *MailgunTransport) Validate(m Message) (err error) {
	if m.Templates.Subject == "" {
		return errors.New("Subject is empty")
	}
	if m.Templates.Text == "" {
		return errors.New("Text is empty")
	}
	u, err := url.Parse(m.Recipient)
	if err == nil {
		_, err = mail.ParseAddress(u.Opaque)
	}
	return
}

func (t *MailgunTransport) Send(m Message) (id string, err error) {

	var subjectBuffer, textBuffer bytes.Buffer

	subjectTemplate := template.Must(template.New("subject").Parse(m.Templates.Subject))
	subjectTemplate.Execute(&subjectBuffer, m.Data)

	textTemplate := template.Must(template.New("text").Parse(m.Templates.Text))
	textTemplate.Execute(&textBuffer, m.Data)

	u, _ := url.Parse(m.Recipient)

	msg := t.mailgun.NewMessage(t.From, subjectBuffer.String(), textBuffer.String(), u.Opaque)
	if t.testMode {
		msg.EnableTestMode()
	}

	for _, f := range m.Attachments {
		msg.AddInline(f)
	}

	if m.Templates.HTML != "" {
		var htmlBuffer bytes.Buffer
		htmlTemplate := template.Must(template.New("html").Parse(m.Templates.HTML))
		htmlTemplate.Execute(&htmlBuffer, m.Data)
		msg.SetHtml(htmlBuffer.String())
	}

	_, id, err = t.mailgun.Send(msg)

	return

}
