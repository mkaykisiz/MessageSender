package service

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/mkaykisiz/sender"
	envvars "github.com/mkaykisiz/sender/configs/env-vars"
	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	redisstore "github.com/mkaykisiz/sender/internal/store/redis"
)

// compile-time proofs of service interface implementation
var _ sender.Service = (*Service)(nil)

// Service represents service
type Service struct {
	l          log.Logger
	ms         mongostore.Store
	rs         redisstore.Store
	envConfigs envvars.Configs
	env        string
	worker     *Worker
}

// NewService creates and returns service
func NewService(l log.Logger, ms mongostore.Store, rs redisstore.Store, envc envvars.Configs, env string, worker *Worker) sender.Service {
	return &Service{
		l:          l,
		ms:         ms,
		rs:         rs,
		envConfigs: envc,
		env:        env,
		worker:     worker,
	}
}

// Health checks health
// swagger:operation GET /health Sender healthRequest
// ---
// summary: Health
// description: checks health
// responses:
//
//	200:
//	  $ref: "#/responses/healthResponse"
func (s *Service) Health(_ context.Context, _ sender.HealthRequest) sender.HealthResponse {
	return sender.HealthResponse{}
}

// StartStopMessageSending starts or stops message sending based on action
// swagger:operation POST /start-stop-sending Sender startStopMessageSendingRequest
// ---
// summary: StartStopMessageSending
// description: starts or stops message sending
// responses:
//
//	  200:
//		  $ref: "#/responses/startStopMessageSendingResponse"
func (s *Service) StartStopMessageSending(_ context.Context, req sender.StartStopMessageSendingRequest) sender.StartStopMessageSendingResponse {
	if req.Action == "start" {
		s.worker.Start()
		return sender.StartStopMessageSendingResponse{Status: "started"}
	} else if req.Action == "stop" {
		s.worker.Stop()
		return sender.StartStopMessageSendingResponse{Status: "stopped"}
	}
	return sender.StartStopMessageSendingResponse{}
}

// RetrieveSentMessages retrieves sent messages
// swagger:operation GET /retrieve-sent-messages Sender retrieveSentMessagesRequest
// ---
// summary: RetrieveSentMessages
// description: retrieves sent messages
// responses:
//
//	  200:
//		  $ref: "#/responses/retrieveSentMessagesResponse"
func (s *Service) RetrieveSentMessages(ctx context.Context, _ sender.RetrieveSentMessagesRequest) sender.RetrieveSentMessagesResponse {
	messages, err := s.ms.GetMessages(ctx, mongostore.MessageFilter{Status: []string{mongostore.STATUS_SENT}}, mongostore.MessageOptions{})
	if err != nil {
		s.log(ctx, err, map[string]interface{}{"method": "RetrieveSentMessages"})
		return sender.RetrieveSentMessagesResponse{}
	}
	return sender.RetrieveSentMessagesResponse{Messages: messages}
}

func (s *Service) StartSendMessage(count int, delay time.Duration) {
	s.worker.Start()
}

func (s *Service) log(ctx context.Context, err error, additionalParams map[string]interface{}) {
	logParams := make([]interface{}, 0, 2+len(additionalParams)*2)

	for k, v := range additionalParams {
		logParams = append(logParams, k, v)
	}

	logParams = append(logParams, "error", err.Error())

	_ = level.Error(s.l).Log(logParams...)
}
