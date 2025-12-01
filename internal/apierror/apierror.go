package apierror

import (
	"net/http"

	"github.com/mkaykisiz/sender/internal/localization"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// error codes
const (
	CodeInternalServerError = 1
	CodeValidationError     = 2
	CodeBadRequestError     = 3
	CodeUnauthorizedError   = 4
)

// error names
const (
	NameInternalServerError = "InternalServerError"
	NameValidationError     = "ValidationError"
	NameUnauthorizedError   = "UnauthorizedError"
	NameBadRequestError     = "BadRequestError"
)

// error actions
const (
	ErrorActionLogout = 100
)

// compile-time proof of error interface implementation
var _ error = (*APIError)(nil)

// compile-time proofs of localizer interface implementation
var _ localization.Localizer = (*APIError)(nil)

// APIError represents api error
type APIError struct {
	Message     string `json:"message"`
	Name        string `json:"name"`
	Code        int    `json:"code"`
	StatusCode  int    `json:"statusCode"`
	ErrorAction int    `json:"errorAction,omitempty"`
	BaseError   error  `json:"-"`

	MessageLocalizerKey string `json:"-"`
}

// Popup represents api error popup
type Popup struct {
	Title          string       `json:"title,omitempty"`
	Message        string       `json:"message"`
	IconURL        string       `json:"iconUrl,omitempty"`
	PositiveButton *PopupButton `json:"positiveButton"`
	NegativeButton *PopupButton `json:"negativeButton"`

	TitleLocalizerKey   string `json:"-"`
	MessageLocalizerKey string `json:"-"`
}

// PopupButton represents api error popup button
type PopupButton struct {
	Text string `json:"text"`

	TextLocalizerKey string `json:"-"`
}

// ToastItem represents api error toast item
type ToastItem struct {
	Message string `json:"message"`
	IconURL string `json:"iconUrl,omitempty"`

	MessageLocalizerKey string `json:"-"`
}

// DefaultInternalServerError represents default internal server error
var DefaultInternalServerError = &APIError{
	Name:                NameInternalServerError,
	Code:                CodeInternalServerError,
	StatusCode:          http.StatusInternalServerError,
	MessageLocalizerKey: "default-internal-server-error-message",
}

// DefaultUnauthorizedError represents default unauthorized error
var DefaultUnauthorizedError = &APIError{
	Name:                NameUnauthorizedError,
	Code:                CodeUnauthorizedError,
	StatusCode:          http.StatusUnauthorized,
	ErrorAction:         ErrorActionLogout,
	MessageLocalizerKey: "default-unauthorized-error-message",
}

// NewValidationError returns validation error
func NewValidationError(message string, messageLocalizerKey string) *APIError {
	return &APIError{
		Message:             message,
		Name:                NameValidationError,
		Code:                CodeValidationError,
		StatusCode:          http.StatusBadRequest,
		MessageLocalizerKey: messageLocalizerKey,
	}
}

// NewBadRequestError returns bad request error
func NewBadRequestError(message string, messageLocalizerKey string) *APIError {
	return &APIError{
		Message:             message,
		Name:                NameBadRequestError,
		Code:                CodeBadRequestError,
		StatusCode:          http.StatusBadRequest,
		MessageLocalizerKey: messageLocalizerKey,
	}
}

// Error returns api error's error message
func (apiErr *APIError) Error() string {
	return apiErr.Message
}

// Localize localizes api error
func (apiErr *APIError) Localize(l *i18n.Localizer) {
	apiErr.Message = localization.Localize(l, apiErr.MessageLocalizerKey, apiErr.Message)
}
