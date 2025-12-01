// Package docs Sender Service API.
//
// Documentation for Sender Service API
//
//	Schemes: https, http
//	BasePath: ./
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//	Security:
//	- api_key:
//
//	SecurityDefinitions:
//	api_key:
//	     type: apiKey
//	     name: KEY
//	     in: header
//
// swagger:meta
package docs

import (
	"github.com/mkaykisiz/sender"
)

type apiError struct {
	Message             string              `json:"message"`
	Name                string              `json:"name"`
	Code                int                 `json:"code"`
	StatusCode          int                 `json:"statusCode"`
	BaseError           error               `json:"baseError"`
	Popup               apiErrorPopup       `json:"popup"`
	ToastItems          []apiErrorToastItem `json:"toastItems"`
	MessageLocalizerKey string              `json:"-"`
}

type apiErrorPopup struct {
	Title               string              `json:"title"`
	Message             string              `json:"message"`
	IconURL             string              `json:"iconUrl"`
	PositiveButton      apiErrorPopupButton `json:"positiveButton"`
	NegativeButton      apiErrorPopupButton `json:"negativeButton"`
	TitleLocalizerKey   string              `json:"-"`
	MessageLocalizerKey string              `json:"-"`
}

type apiErrorPopupButton struct {
	Text             string `json:"text"`
	TextLocalizerKey string `json:"-"`
}

type apiErrorToastItem struct {
	Message             string `json:"message"`
	IconURL             string `json:"iconUrl"`
	MessageLocalizerKey string `json:"-"`
}

// swagger:parameters request-header
type requestHeader struct {
	// in: header
	// name: Accept-Language
	// example: TR
	// default: tr
	AcceptLanguage string `json:"Accept-Language"`
}

// swagger:parameters startStopMessageSendingRequest
type startStopMessageSendingRequest struct {
	requestHeader
	// in: body
	Body struct {
		// required: true
		// enum: ["start", "stop"]
		Action string `json:"action"`
	}
}

// Success
// swagger:response startStopMessageSendingResponse
type startStopMessageSendingResponse struct {
	Body struct {
		Status string    `json:"status"`
		Result *apiError `json:"result"`
	}
}

// swagger:parameters retrieveSentMessagesRequest
type retrieveSentMessagesRequest struct {
	requestHeader
}

// Success
// swagger:response retrieveSentMessagesResponse
type retrieveSentMessagesResponse struct {
	Body struct {
		Messages []sender.MessageTransaction `json:"messages"`
		Result   *apiError                   `json:"result"`
	}
}
