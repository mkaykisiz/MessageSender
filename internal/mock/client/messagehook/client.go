package messagehook

import (
	"context"
	messagehookclient "github.com/mkaykisiz/sender/internal/client/messageclient"

	"github.com/stretchr/testify/mock"
)

// compile-time proof of interface implementation
var _ messagehookclient.MessageClient = (*Client)(nil)

// Client represents mock client
type Client struct {
	mock.Mock
}

// NewClient returns mock client
func NewClient() *Client {
	return &Client{}
}

// SendMessage mocks send message method
func (c *Client) SendMessage(ctx context.Context, to, content string) (*messagehookclient.MessageResponse, error) {
	args := c.Called(ctx, to, content)

	return args.Get(0).(*messagehookclient.MessageResponse), args.Error(1)
}
