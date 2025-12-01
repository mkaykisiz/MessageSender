package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/mkaykisiz/sender"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/iris-contrib/schema"
	"github.com/mkaykisiz/sender/internal/apierror"
	"github.com/mkaykisiz/sender/internal/endpoints"
	"github.com/mkaykisiz/sender/internal/localization"
	"github.com/mkaykisiz/sender/internal/transport"
)

// endpoint names
const (
	health                  = "Health"
	startStopMessageSending = "StartStopMessageSending"
	retrieveSentMessages    = "RetrieveSentMessages"
)

// decoder tags
const (
	headerTag = "header"
	queryTag  = "query"
)

const invalidResponseError = "invalid response"
const multipartFormSizeLimit = 10 * 1024 * 1024

// MakeHTTPHandler makes and returns http handler
func MakeHTTPHandler(l log.Logger, s sender.Service) http.Handler {
	es := endpoints.MakeEndpoints(s)

	r := mux.NewRouter()

	// health GET /health
	r.Methods("GET").Path("/health").Handler(
		makeHealthHandler(es.HealthEndpoint, makeDefaultServerOptions(l, health)),
	)

	// start-stop-message-sending POST /start-stop-sending
	r.Methods("POST").Path("/start-stop-sending").Handler(
		makeStartStopMessageSendingHandler(es.StartStopMessageSendingEndpoint, makeDefaultServerOptions(l, startStopMessageSending)),
	)

	// retrieve-sent-messages GET /retrieve-sent-messages
	r.Methods("GET").Path("/retrieve-sent-messages").Handler(
		makeRetrieveSentMessagesHandler(es.RetrieveSentMessagesEndpoint, makeDefaultServerOptions(l, retrieveSentMessages)),
	)

	// core services docs
	swaggerRouter := r.PathPrefix("/docs").Subrouter()

	// replies to the request with the contents of the named swagger.yml
	swaggerRouter.
		Methods(http.MethodGet).
		Path("/swagger.yaml").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.yaml")
		})

	// responds to an HTTP swagger request
	swaggerRouter.
		Methods(http.MethodGet).
		Path("").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opts := middleware.SwaggerUIOpts{
				SpecURL: "/docs/swagger.yaml",
				Path:    "/docs"}
			middleware.SwaggerUI(opts, nil).ServeHTTP(w, r)
		})

	return r
}

func makeHealthHandler(e endpoint.Endpoint, serverOptions []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(sender.HealthRequest{}), encoder, serverOptions...)
	return h
}

func makeStartStopMessageSendingHandler(e endpoint.Endpoint, serverOptions []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(sender.StartStopMessageSendingRequest{}), encoder, serverOptions...)
	return h
}

func makeRetrieveSentMessagesHandler(e endpoint.Endpoint, serverOptions []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(sender.RetrieveSentMessagesRequest{}), encoder, serverOptions...)
	return h
}

func makeDefaultServerOptions(l log.Logger, endpointName string) []kithttp.ServerOption {
	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(errorEncoder),
		kithttp.ServerErrorHandler(transport.NewErrorHandler(l, endpointName)),
		kithttp.ServerBefore(localization.AddLocalizerToContext),
	}
	return options
}

func makeDecoder(emptyReq interface{}) kithttp.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		req := reflect.New(reflect.TypeOf(emptyReq)).Interface()

		req.(sender.Request).SetIPAddress(getIPAddress(r))

		if err := newHeaderDecoder().Decode(req, r.Header); err != nil {
			return nil, fmt.Errorf("decoding request header failed, %s", err.Error())
		}

		if err := newQueryDecoder().Decode(req, r.URL.Query()); err != nil {
			return nil, fmt.Errorf("decoding request query failed, %s", err.Error())
		}

		if requestHasBody(r) {
			formValueTags := getFormValueTags(req)
			formFileTags := getFormFileTags(req)
			requestHasFormData := len(formValueTags) > 0 || len(formFileTags) > 0
			if requestHasFormData {
				if err := r.ParseMultipartForm(multipartFormSizeLimit); err != nil {
					return nil, fmt.Errorf("parsing multipart form failed, %s", err.Error())
				}

				for _, tag := range formValueTags {
					value := r.FormValue(tag)

					setFormValue(tag, value, req)
				}

				for _, tag := range formFileTags {
					f, h, err := r.FormFile(tag)
					if err != nil {
						return nil, fmt.Errorf("getting multipart form file failed, %s", err.Error())
					}

					value := make([]byte, h.Size)
					_, err = f.Read(value)
					if err != nil {
						return nil, fmt.Errorf("reading multipart form file failed, %s", err.Error())
					}

					setFormFile(tag, value, req)
				}
			} else {
				if err := json.NewDecoder(r.Body).Decode(req); err != nil {
					return nil, fmt.Errorf("decoding request body failed, %s", err.Error())
				}
			}
		}

		if err := validate(req); err != nil {
			apiError := apierror.NewValidationError(err.Error(), "")
			apiError.BaseError = err
			return nil, apiError
		}

		return req, nil
	}
}

func getIPAddress(r *http.Request) string {
	if ipAddress := r.Header.Get("X-Real-Ip"); ipAddress != "" {
		return ipAddress
	}

	if ipAddress := r.Header.Get("X-Forwarded-For"); ipAddress != "" {
		return ipAddress
	}

	return r.RemoteAddr
}

func newHeaderDecoder() *schema.Decoder {
	return newDecoder(headerTag)
}

func newQueryDecoder() *schema.Decoder {
	return newDecoder(queryTag)
}

func newDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if tag != "" {
		decoder.SetAliasTag(tag)
	}

	return decoder
}

func requestHasBody(r *http.Request) bool {
	return r.Body != http.NoBody
}

func getFormValueTags(req interface{}) []string {
	return getTags("form-value", req)
}

func getFormFileTags(req interface{}) []string {
	return getTags("form-file", req)
}

func getTags(tagName string, req interface{}) []string {
	e := reflect.ValueOf(req).Elem()

	tt := make([]string, 0)
	for i := 0; i < e.NumField(); i++ {
		tf := e.Type().Field(i)

		t := tf.Tag.Get(tagName)
		if t == "" || t == "-" {
			continue
		}

		tt = append(tt, t)
	}

	return tt
}

func setFormValue(tag string, value interface{}, req interface{}) {
	setValue("form-value", tag, value, req)
}

func setFormFile(tag string, value interface{}, req interface{}) {
	setValue("form-file", tag, value, req)
}

func setValue(tagName string, tag string, value interface{}, req interface{}) {
	e := reflect.ValueOf(req).Elem()

	for i := 0; i < e.NumField(); i++ {
		tf := e.Type().Field(i)

		t := tf.Tag.Get(tagName)
		if t != tag {
			continue
		}

		e.Field(i).Set(reflect.ValueOf(value))
	}
}

func validate(req interface{}) error {
	errs := validator.New().Struct(req)
	if errs == nil {
		return nil
	}

	firstErr := errs.(validator.ValidationErrors)[0]

	return fmt.Errorf("validation failed, tag: %s, field: %s", firstErr.Tag(), firstErr.Field())
}

func encoder(ctx context.Context, rw http.ResponseWriter, response interface{}) error {

	r, ok := response.(sender.Response)
	if !ok {
		return errors.New(invalidResponseError)
	}

	if r.APIError() != nil {

		errorEncoder(ctx, r.APIError(), rw)
		return nil
	}

	l := localization.GetLocalizerFromContext(ctx)
	lr := r.Localize(l)

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	return json.NewEncoder(rw).Encode(lr)
}

type errorResponse struct {
	Data   interface{}        `json:"data"`
	Result *apierror.APIError `json:"result"`
}

func errorEncoder(ctx context.Context, err error, rw http.ResponseWriter) {

	apiErr, ok := err.(*apierror.APIError)
	if !ok {
		apiErr = apierror.DefaultInternalServerError
	}

	l := localization.GetLocalizerFromContext(ctx)
	apiErr.Localize(l)

	er := errorResponse{
		Data:   nil,
		Result: apiErr,
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(apiErr.StatusCode)
	_ = json.NewEncoder(rw).Encode(er)
}
