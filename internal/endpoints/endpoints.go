package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mkaykisiz/sender"
)

// Endpoints represents service endpoints
type Endpoints struct {
	HealthEndpoint                  endpoint.Endpoint
	StartStopMessageSendingEndpoint endpoint.Endpoint
	RetrieveSentMessagesEndpoint    endpoint.Endpoint
}

// MakeEndpoints makes and returns endpoints
func MakeEndpoints(s sender.Service) Endpoints {
	return Endpoints{
		HealthEndpoint:                  MakeHealthEndpoint(s),
		StartStopMessageSendingEndpoint: MakeStartStopMessageSendingEndpoint(s),
		RetrieveSentMessagesEndpoint:    MakeRetrieveSentMessagesEndpoint(s),
	}
}

// MakeHealthEndpoint makes and returns health endpoint
func MakeHealthEndpoint(s sender.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*sender.HealthRequest)

		res := s.Health(ctx, *req)

		return res, nil
	}
}

// MakeStartStopMessageSendingEndpoint makes and returns start stop message sending endpoint
func MakeStartStopMessageSendingEndpoint(s sender.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*sender.StartStopMessageSendingRequest)

		res := s.StartStopMessageSending(ctx, *req)

		return res, nil
	}
}

// MakeRetrieveSentMessagesEndpoint makes and returns retrieve sent messages endpoint
func MakeRetrieveSentMessagesEndpoint(s sender.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*sender.RetrieveSentMessagesRequest)

		res := s.RetrieveSentMessages(ctx, *req)

		return res, nil
	}
}
