package messageclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)
var ClientHost = "http://localhost:8080"
var ClientAuthKey = "test-key"
var PhoneNumber = "+905551234567"
var Message = "Test message"

func TestNewClient(t *testing.T) {
	t.Run("with default http client", func(t *testing.T) {
		client := NewClient(ClientHost, ClientAuthKey, 3, 1*time.Second)

		assert.NotNil(t, client)
		assert.Equal(t, ClientHost, client.url)
		assert.Equal(t, ClientAuthKey, client.authKey)
		assert.Equal(t, 3, client.maxRetries)
		assert.Equal(t, 1*time.Second, client.retryDelay)
		assert.Equal(t, http.DefaultClient, client.c)
	})
}

func TestMessageClient_SendMessage(t *testing.T) {
	t.Run("successful send with 200 OK", func(t *testing.T) {
		expectedResponse := MessageResponse{
			Message:   "Message sent successfully",
			MessageID: "msg-123",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			assert.Equal(t, http.MethodPost, r.Method)

			// Verify headers
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, ClientAuthKey, r.Header.Get("x-ins-auth-key"))

			// Verify request body
			var req MessageRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, PhoneNumber, req.To)
			assert.Equal(t, Message, req.Content)

			// Send response
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedResponse.Message, response.Message)
		assert.Equal(t, expectedResponse.MessageID, response.MessageID)
	})

	t.Run("successful send with 202 Accepted", func(t *testing.T) {
		expectedResponse := MessageResponse{
			Message:   "Message accepted",
			MessageID: "msg-456",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedResponse.MessageID, response.MessageID)
	})

	t.Run("failed send with 400 Bad Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("failed send with 500 Internal Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("failed send with invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "decoding")
	})

	t.Run("failed send with network error", func(t *testing.T) {
		client := NewClient("http://invalid-url-that-does-not-exist.local", ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(MessageResponse{MessageID: "msg-789"})
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		response, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("verify auth key is sent correctly", func(t *testing.T) {
		var receivedAuthKey string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuthKey = r.Header.Get("x-ins-auth-key")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(MessageResponse{MessageID: "msg-999"})
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		_, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.NoError(t, err)
		assert.Equal(t, ClientAuthKey, receivedAuthKey)
	})

	t.Run("verify content-type header", func(t *testing.T) {
		var receivedContentType string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedContentType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(MessageResponse{MessageID: "msg-888"})
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		_, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.NoError(t, err)
		assert.Equal(t, "application/json", receivedContentType)
	})

	t.Run("verify request payload", func(t *testing.T) {
		var receivedRequest MessageRequest

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&receivedRequest)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(MessageResponse{MessageID: "msg-777"})
		}))
		defer server.Close()

		client := NewClient(server.URL, ClientAuthKey, 3, 1*time.Second)
		ctx := context.Background()

		_, err := client.SendMessage(ctx, PhoneNumber, Message)

		assert.NoError(t, err)
		assert.Equal(t, PhoneNumber, receivedRequest.To)
		assert.Equal(t, Message, receivedRequest.Content)
	})
}
