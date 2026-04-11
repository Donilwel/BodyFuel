package sendgrid

import (
	"fmt"

	sendgridgo "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Config struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type Client struct {
	client    *sendgridgo.Client
	fromEmail string
	fromName  string
}

func NewClient(cfg Config) *Client {
	return &Client{
		client:    sendgridgo.NewSendClient(cfg.APIKey),
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

func (c *Client) SendEmail(to, subject, body string) error {
	from := mail.NewEmail(c.fromName, c.fromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	resp, err := c.client.Send(message)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: status %d: %s", resp.StatusCode, resp.Body)
	}

	return nil
}
