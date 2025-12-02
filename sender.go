package sender

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	apierror "github.com/mkaykisiz/sender/internal/apierror"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// service environments
const (
	Local = "local"
	Dev   = "dev"
	Stg   = "stg"
	Prod  = "prod"
)

// language codes
const (
	LanguageCodeTR = "tr"
	LanguageCodeEN = "en"
)

const (
	MaxMessageLength = 1000
)

var (
	LanguageCodes = []string{LanguageCodeTR, LanguageCodeEN}
)

type (
	ResponseMessage struct {
		ID        string     `json:"id"`
		Content   string     `json:"content"`
		Recipient string     `json:"recipient"`
		Status    string     `json:"status"` // "pending", "sent", "failed"
		SentAt    *time.Time `json:"sent_at,omitempty"`
	}

	MessageTransaction struct {
		ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
		Content   string             `json:"content" bson:"content" validate:"required,max=1000"`
		Recipient string             `json:"recipient" bson:"recipient"` // TODO birden fazla adi var
		Status    string             `json:"status" bson:"status"`       // "pending", "sent", "failed"
		SentAt    *time.Time         `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
		CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	}
)

func (m *MessageTransaction) IsValid() bool {
	if len(m.Content) > MaxMessageLength || len(m.Recipient) == 0 || len(m.Content) == 0 {
		return false
	}
	return true
}

type HealthStatus atomic.Bool

func (s *HealthStatus) SetStatus(state bool) {
	val := (*atomic.Bool)(s)
	val.Store(state)
}

func (s *HealthStatus) GetStatus() bool {
	val := (*atomic.Bool)(s)
	return val.Load()
}

var (
	HEALTH_STATUS         HealthStatus
	ErrServiceUnavailable = errors.New("service unavailable")
)

// Service defines behaviors of sample service
type Service interface {
	Health(context.Context, HealthRequest) HealthResponse
	StartStopMessageSending(context.Context, StartStopMessageSendingRequest) StartStopMessageSendingResponse
	RetrieveSentMessages(context.Context, RetrieveSentMessagesRequest) RetrieveSentMessagesResponse

	StartSendMessage(count int, delay time.Duration)
}

// Request defines behaviors of request
type Request interface {
	SetIPAddress(ipAddress string)
}

// Response defines behaviors of response
type Response interface {
	Localize(l *i18n.Localizer) interface{}
	APIError() error
}

// compile-time proofs of request interface implementation
var (
	_ Request = (*HealthRequest)(nil)
)

// compile-time proofs of response interface implementation
var (
	_ Response = (*HealthResponse)(nil)
)

// HealthRequest and HealthResponse represents health request and response
type (
	HealthRequest  struct{}
	HealthResponse struct{}
)

// StartStopMessageSendingRequest and StartStopMessageSendingResponse represents request and response
type (
	StartStopMessageSendingRequest struct {
		IPAddress string `json:"-"`
		Action    string `json:"action" validate:"required,oneof=start stop"`
	}
	StartStopMessageSendingResponse struct {
		Result *apierror.APIError `json:"result"`
		Status string             `json:"status"` // "started" or "stopped"
	}
)

// RetrieveSentMessagesRequest and RetrieveSentMessagesResponse represents request and response
type (
	RetrieveSentMessagesRequest struct {
		IPAddress string `json:"-"`
	}
	RetrieveSentMessagesResponse struct {
		Result   *apierror.APIError   `json:"result"`
		Messages []MessageTransaction `json:"messages"`
	}
)

// Header represents header
type Header struct {
	AcceptLanguage string `json:"-" header:"Accept-Language"`
}

// SetIPAddress does nothing since health request doesn't have ip address
func (r *HealthRequest) SetIPAddress(_ string) {}

// SetIPAddress request's ip address
func (r *StartStopMessageSendingRequest) SetIPAddress(ipAddress string) {
	r.IPAddress = ipAddress
}

// SetIPAddress request's ip address
func (r *RetrieveSentMessagesRequest) SetIPAddress(ipAddress string) {
	r.IPAddress = ipAddress
}

// APIError returns error when API is shutting down
func (r HealthResponse) APIError() error {
	if !HEALTH_STATUS.GetStatus() {
		return ErrServiceUnavailable
	}
	return nil
}

// APIError returns error when API is shutting down
func (r StartStopMessageSendingResponse) APIError() error {
	if r.Result == nil {
		return nil
	}

	return r.Result
}

// APIError returns error when API is shutting down
func (r RetrieveSentMessagesResponse) APIError() error {
	if r.Result == nil {
		return nil
	}

	return r.Result
}

// Localize localizes response
func (r HealthResponse) Localize(_ *i18n.Localizer) interface{} {
	return r
}

// Localize localizes response
func (r StartStopMessageSendingResponse) Localize(_ *i18n.Localizer) interface{} {
	return r
}

// Localize localizes response
func (r RetrieveSentMessagesResponse) Localize(_ *i18n.Localizer) interface{} {
	return r
}
