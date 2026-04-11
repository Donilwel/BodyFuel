package apns

import (
	"encoding/json"
	"fmt"

	apns2 "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
)

type Config struct {
	KeyPath  string
	KeyID    string
	TeamID   string
	BundleID string
	Sandbox  bool
}

type Payload struct {
	Title string
	Body  string
}

type Client struct {
	client   *apns2.Client
	bundleID string
}

func NewClient(cfg Config) (*Client, error) {
	authKey, err := token.AuthKeyFromFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load apns auth key: %w", err)
	}

	t := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}

	client := apns2.NewTokenClient(t)
	if cfg.Sandbox {
		client = client.Development()
	} else {
		client = client.Production()
	}

	return &Client{
		client:   client,
		bundleID: cfg.BundleID,
	}, nil
}

func (c *Client) Send(deviceToken string, p Payload) error {
	payload := map[string]any{
		"aps": map[string]any{
			"alert": map[string]string{
				"title": p.Title,
				"body":  p.Body,
			},
			"sound": "default",
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     raw,
	}

	resp, err := c.client.Push(notification)
	if err != nil {
		return fmt.Errorf("push notification: %w", err)
	}

	if !resp.Sent() {
		return fmt.Errorf("apns error: %s (%d)", resp.Reason, resp.StatusCode)
	}

	return nil
}
