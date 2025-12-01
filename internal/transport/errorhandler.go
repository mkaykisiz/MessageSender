package transport

import (
	"context"

	"github.com/go-kit/kit/log"
	kittransport "github.com/go-kit/kit/transport"
)

// compile-time proof of go-kit transport's error handler interface implementation
var _ kittransport.ErrorHandler = (*ErrorHandler)(nil)

// ErrorHandler represents error handler
type ErrorHandler struct {
	l            log.Logger
	EndpointName string
}

// NewErrorHandler creates and returns error handler
func NewErrorHandler(l log.Logger, endpointName string) *ErrorHandler {
	return &ErrorHandler{
		l:            l,
		EndpointName: endpointName,
	}
}

// Handle logs error
func (eh *ErrorHandler) Handle(_ context.Context, err error) {
	_ = eh.l.Log("error", err.Error())
}
