package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/mkaykisiz/sender"
	envvars "github.com/mkaykisiz/sender/configs/env-vars"
	mockmessagehook "github.com/mkaykisiz/sender/internal/mock/client/messagehook"
	mockmongostore "github.com/mkaykisiz/sender/internal/mock/store/mongo"
	mockredisstore "github.com/mkaykisiz/sender/internal/mock/store/redis"
	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestService_RetrieveSentMessages(t *testing.T) {
	mockMongoStore := mockmongostore.NewStore()
	mockRedisStore := mockredisstore.NewStore()
	mockMessageClient := mockmessagehook.NewClient()
	logger := log.NewNopLogger()

	worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)
	svc := NewService(logger, mockMongoStore, mockRedisStore, envvars.Configs{}, "test", worker)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedMessages := []sender.MessageTransaction{
			{ID: primitive.NewObjectID(), Content: "test", Status: "sent"},
		}

		mockMongoStore.On("GetMessages", ctx, mongostore.MessageFilter{Status: []string{mongostore.STATUS_SENT}}, mongostore.MessageOptions{}).Return(expectedMessages, nil).Once()

		resp := svc.RetrieveSentMessages(ctx, sender.RetrieveSentMessagesRequest{})

		assert.Nil(t, resp.Result)
		assert.Equal(t, expectedMessages, resp.Messages)
		mockMongoStore.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockMongoStore.On("GetMessages", ctx, mongostore.MessageFilter{Status: []string{mongostore.STATUS_SENT}}, mongostore.MessageOptions{}).Return([]sender.MessageTransaction(nil), errors.New("db error")).Once()

		resp := svc.RetrieveSentMessages(ctx, sender.RetrieveSentMessagesRequest{})

		assert.Nil(t, resp.Result)
		assert.Empty(t, resp.Messages)
		mockMongoStore.AssertExpectations(t)
	})
}

func TestService_StartStopMessageSending(t *testing.T) {
	t.Run("start action", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)
		svc := NewService(logger, mockMongoStore, mockRedisStore, envvars.Configs{}, "test", worker)
		ctx := context.Background()

		mockMongoStore.On("GetMessages", mock.Anything, mock.Anything, mock.Anything).Return([]sender.MessageTransaction{}, nil).Maybe()

		resp := svc.StartStopMessageSending(ctx, sender.StartStopMessageSendingRequest{Action: "start"})
		
		assert.Nil(t, resp.Result)
		assert.Equal(t, "started", resp.Status)
		
		// Stop to cleanup
		svc.StartStopMessageSending(ctx, sender.StartStopMessageSendingRequest{Action: "stop"})
	})

	t.Run("stop action", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)
		svc := NewService(logger, mockMongoStore, mockRedisStore, envvars.Configs{}, "test", worker)
		ctx := context.Background()

		resp := svc.StartStopMessageSending(ctx, sender.StartStopMessageSendingRequest{Action: "stop"})
		
		assert.Nil(t, resp.Result)
		assert.Equal(t, "stopped", resp.Status)
	})

	t.Run("invalid action", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)
		svc := NewService(logger, mockMongoStore, mockRedisStore, envvars.Configs{}, "test", worker)
		ctx := context.Background()

		resp := svc.StartStopMessageSending(ctx, sender.StartStopMessageSendingRequest{Action: "invalid"})
		
		assert.Nil(t, resp.Result)
		assert.Empty(t, resp.Status)
	})
}

func TestService_StartSendMessage(t *testing.T) {
	mockMongoStore := mockmongostore.NewStore()
	mockRedisStore := mockredisstore.NewStore()
	mockMessageClient := mockmessagehook.NewClient()
	logger := log.NewNopLogger()

	worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)
	svc := NewService(logger, mockMongoStore, mockRedisStore, envvars.Configs{}, "test", worker)

	mockMongoStore.On("GetMessages", mock.Anything, mock.Anything, mock.Anything).Return([]sender.MessageTransaction{}, nil).Maybe()

	// Just verify it doesn't panic
	svc.StartSendMessage(2, 1*time.Second)

	worker.Stop()
}
