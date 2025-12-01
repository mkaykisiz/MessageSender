package messageclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MessageRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type MessageResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

// MessageClient defines behaviors of message client
type MessageClient interface {
	SendMessage(ctx context.Context, to, content string) (*MessageResponse, error)
}

type messageClient struct {
	url        string
	authKey    string
	maxRetries int
	retryDelay time.Duration
	c          *http.Client
}

// NewClient creates and returns client
func NewClient(url, authKey string, maxRetries int, retryDelay time.Duration) *messageClient {
	cli := &messageClient{
		url:        url,
		authKey:    authKey,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		c:          http.DefaultClient,
	}

	return cli
}

// SendMessage returns sent message response
func (c *messageClient) SendMessage(ctx context.Context, to, content string) (*MessageResponse, error) {
	hookRes := MessageResponse{}

	payload := MessageRequest{
		To:      to,
		Content: content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("getting support url config failed while creating http request, %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", c.authKey)

	res, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting support url config failed while doing http request, %s", err.Error())
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("getting support url config failed while doing http request, %s", res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&hookRes)
	if err != nil {
		return nil, fmt.Errorf("getting support url config failed while decoding support config, %s", err.Error())
	}

	return &hookRes, nil
}
