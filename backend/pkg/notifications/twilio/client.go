package twilio

import (
	"fmt"

	twiliolib "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Config struct {
	AccountSID string
	AuthToken  string
	FromPhone  string
}

type Client struct {
	client    *twiliolib.RestClient
	fromPhone string
}

func NewClient(cfg Config) *Client {
	return &Client{
		client: twiliolib.NewRestClientWithParams(twiliolib.ClientParams{
			Username: cfg.AccountSID,
			Password: cfg.AuthToken,
		}),
		fromPhone: cfg.FromPhone,
	}
}

func (c *Client) SendSMS(to, body string) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(c.fromPhone)
	params.SetBody(body)

	_, err := c.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	return nil
}
