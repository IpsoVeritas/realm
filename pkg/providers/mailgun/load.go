package mailgun

import (
	"errors"

	realm "github.com/Brickchain/realm"
	messaging "github.com/Brickchain/realm/pkg/providers/messaging"
	"github.com/spf13/viper"
)

type MailgunProvider struct {
	t *messaging.MailgunTransport
}

func NewMailgunProvider(configFile string) (realm.EmailProvider, error) {
	config := viper.New()
	config.AddConfigPath("./")
	config.AddConfigPath("/")
	config.SetConfigFile(configFile)

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	if !config.IsSet("mailgun") {
		return nil, errors.New("No mailgun configuration found")
	}
	mailgunConfig := config.Sub("mailgun")

	return &MailgunProvider{
		t: messaging.NewMailgunTransport(mailgunConfig),
	}, nil
}

func (p *MailgunProvider) Send(msg messaging.Message) (*realm.EmailStatus, error) {
	status := &realm.EmailStatus{}

	msgID, err := p.t.Send(msg)
	status.MessageID = msgID
	if err != nil {
		status.Sent = false
		return status, err
	}

	status.Sent = true

	return status, nil
}

func (p *MailgunProvider) Validate(msg messaging.Message) error {
	return p.t.Validate(msg)
}
