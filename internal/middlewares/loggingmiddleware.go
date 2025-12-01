package middlewares

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/mkaykisiz/sender"
	"time"
)

// LoggingMiddleware represents logging middleware
type LoggingMiddleware struct {
	l    log.Logger
	next sender.Service
}

// NewLoggingMiddleware creates and returns logging middleware
func NewLoggingMiddleware(l log.Logger) Middleware {
	return func(next sender.Service) sender.Service {
		return &LoggingMiddleware{
			l:    l,
			next: next,
		}
	}
}

// Health represents logging middleware for Health method
func (m *LoggingMiddleware) Health(ctx context.Context, req sender.HealthRequest) sender.HealthResponse {
	return m.next.Health(ctx, req)
}

// StartStopMessageSending represents logging middleware for StartStopMessageSending method
func (m *LoggingMiddleware) StartStopMessageSending(ctx context.Context, req sender.StartStopMessageSendingRequest) sender.StartStopMessageSendingResponse {
	res := m.next.StartStopMessageSending(ctx, req)
	if res.Result != nil {
		m.logWithLogger(res.Result.BaseError, map[string]interface{}{
			"method":    "StartStopMessageSending",
			"action":    req.Action,
			"ipAddress": req.IPAddress,
		})
	}
	return res
}

// RetrieveSentMessages represents logging middleware for RetrieveSentMessages method
func (m *LoggingMiddleware) RetrieveSentMessages(ctx context.Context, req sender.RetrieveSentMessagesRequest) sender.RetrieveSentMessagesResponse {
	res := m.next.RetrieveSentMessages(ctx, req)
	if res.Result != nil {
		m.logWithLogger(res.Result.BaseError, map[string]interface{}{
			"method":    "RetrieveSentMessages",
			"ipAddress": req.IPAddress,
		})
	}
	return res
}

// StartSendMessage represents logging middleware for StartSendMessage method
func (m *LoggingMiddleware) StartSendMessage(count int, delay time.Duration) {

}

func (m *LoggingMiddleware) logWithLogger(err error, additionalParams map[string]interface{}) {
	logParams := make([]interface{}, 0, 2+len(additionalParams)*2)

	for k, v := range additionalParams {
		logParams = append(logParams, k, v)
	}

	logParams = append(logParams, "error", err.Error())

	_ = m.l.Log(logParams...)
}
