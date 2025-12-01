package service

import (
	"errors"
	"testing"
	"time"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/mkaykisiz/sender"
	"github.com/mkaykisiz/sender/internal/client/messageclient"
	mockmessagehook "github.com/mkaykisiz/sender/internal/mock/client/messagehook"
	mockmongostore "github.com/mkaykisiz/sender/internal/mock/store/mongo"
	mockredisstore "github.com/mkaykisiz/sender/internal/mock/store/redis"
	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWorker_StartStop(t *testing.T) {
	t.Run("start worker", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		mockMongoStore.On("GetMessages", mock.Anything, mock.Anything, mock.Anything).Return([]sender.MessageTransaction{}, nil).Maybe()

		worker.Start()
		
		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)
		
		assert.True(t, worker.running)
		
		worker.Stop()
	})

	t.Run("stop worker", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		mockMongoStore.On("GetMessages", mock.Anything, mock.Anything, mock.Anything).Return([]sender.MessageTransaction{}, nil).Maybe()

		worker.Start()
		time.Sleep(100 * time.Millisecond)
		
		worker.Stop()
		
		// Give it a moment to stop
		time.Sleep(100 * time.Millisecond)
		
		assert.False(t, worker.running)
	})

	t.Run("start already running worker", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		mockMongoStore.On("GetMessages", mock.Anything, mock.Anything, mock.Anything).Return([]sender.MessageTransaction{}, nil).Maybe()

		worker.Start()
		time.Sleep(100 * time.Millisecond)
		
		// Try to start again
		worker.Start()
		
		assert.True(t, worker.running)
		
		worker.Stop()
	})

	t.Run("stop already stopped worker", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()
		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		// Worker should already be stopped
		worker.Stop()
		
		// Try to stop again - should not panic
		worker.Stop()
		
		assert.False(t, worker.running)
	})
}

func TestWorker_Process(t *testing.T) {
	t.Run("process with no messages", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		mockMongoStore.On("GetMessages", mock.Anything, 
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return([]sender.MessageTransaction{}, nil).Once()

		// Call process directly
		worker.process()

		mockMongoStore.AssertExpectations(t)
	})

	t.Run("process with successful message sending", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		msgID := primitive.NewObjectID()
		messages := []sender.MessageTransaction{
			{
				ID:        msgID,
				Content:   "Test message",
				Recipient: "+905551234567",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			},
		}

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

		mockMessageClient.On("SendMessage", mock.Anything, "+905551234567", "Test message").
			Return(&messageclient.MessageResponse{MessageID: msgID.Hex()}, nil).Once()

		mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID, mongostore.STATUS_SENT, mock.Anything).
			Return(nil).Once()

		mockRedisStore.On("CacheMessageID", mock.Anything, msgID.Hex()).
			Return(nil).Once()

		worker.process()

		mockMongoStore.AssertExpectations(t)
		mockMessageClient.AssertExpectations(t)
		mockRedisStore.AssertExpectations(t)
	})

	t.Run("process with failed message sending", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		msgID := primitive.NewObjectID()
		messages := []sender.MessageTransaction{
			{
				ID:        msgID,
				Content:   "Test message",
				Recipient: "+905551234567",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			},
		}

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

		mockMessageClient.On("SendMessage", mock.Anything, "+905551234567", "Test message").
			Return((*messageclient.MessageResponse)(nil), errors.New("send failed")).Once()

		mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID, mongostore.STATUS_FAILED, mock.Anything).
			Return(nil).Once()

		worker.process()

		mockMongoStore.AssertExpectations(t)
		mockMessageClient.AssertExpectations(t)
	})

	t.Run("process with database error on fetch", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return([]sender.MessageTransaction(nil), errors.New("db error")).Once()

		worker.process()

		mockMongoStore.AssertExpectations(t)
	})

	t.Run("process with update status error", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		msgID := primitive.NewObjectID()
		messages := []sender.MessageTransaction{
			{
				ID:        msgID,
				Content:   "Test message",
				Recipient: "+905551234567",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			},
		}

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

		mockMessageClient.On("SendMessage", mock.Anything, "+905551234567", "Test message").
			Return(&messageclient.MessageResponse{}, nil).Once()

		mockMongoStore.On("UpdateMessageStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("update error")).Maybe()

		fmt.Printf("Expected calls: %+v\n", mockMongoStore.ExpectedCalls)

		worker.process()

		mockMongoStore.AssertExpectations(t)
		mockMessageClient.AssertExpectations(t)
	})

	t.Run("process with redis cache error", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		msgID := primitive.NewObjectID()
		messages := []sender.MessageTransaction{
			{
				ID:        msgID,
				Content:   "Test message",
				Recipient: "+905551234567",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			},
		}

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

		mockMessageClient.On("SendMessage", mock.Anything, "+905551234567", "Test message").
			Return(&messageclient.MessageResponse{MessageID: msgID.Hex()}, nil).Once()

		mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID, mongostore.STATUS_SENT, mock.Anything).
			Return(nil).Once()

		mockRedisStore.On("CacheMessageID", mock.Anything, msgID.Hex()).
			Return(errors.New("redis error")).Once()

		worker.process()

		mockMongoStore.AssertExpectations(t)
		mockMessageClient.AssertExpectations(t)
		mockRedisStore.AssertExpectations(t)
	})

	t.Run("process with message exceeding character limit", func(t *testing.T) {
		mockMongoStore := mockmongostore.NewStore()
		mockRedisStore := mockredisstore.NewStore()
		mockMessageClient := mockmessagehook.NewClient()
		logger := log.NewNopLogger()

		worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

		msgID := primitive.NewObjectID()
		longContent := make([]byte, 1001)
		for i := range longContent {
			longContent[i] = 'a'
		}

		messages := []sender.MessageTransaction{
			{
				ID:        msgID,
				Content:   string(longContent),
				Recipient: "+905551234567",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			},
		}

		mockMongoStore.On("GetMessages", mock.Anything,
			mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
			mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

		mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID, mongostore.STATUS_FAILED, mock.Anything).
			Return(nil).Once()

		worker.process()

		mockMongoStore.AssertExpectations(t)
		// Message client should NOT be called
		mockMessageClient.AssertNotCalled(t, "SendMessage")
	})
}

func TestWorker_MultipleMessages(t *testing.T) {
	mockMongoStore := mockmongostore.NewStore()
	mockRedisStore := mockredisstore.NewStore()
	mockMessageClient := mockmessagehook.NewClient()
	logger := log.NewNopLogger()

	worker := NewWorker(mockMessageClient, mockMongoStore, mockRedisStore, logger, 2)

	msgID1 := primitive.NewObjectID()
	msgID2 := primitive.NewObjectID()
	messages := []sender.MessageTransaction{
		{
			ID:        msgID1,
			Content:   "Test message 1",
			Recipient: "+905551234567",
			Status:    mongostore.STATUS_PENDING,
			CreatedAt: time.Now(),
		},
		{
			ID:        msgID2,
			Content:   "Test message 2",
			Recipient: "+905559876543",
			Status:    mongostore.STATUS_PENDING,
			CreatedAt: time.Now(),
		},
	}

	mockMongoStore.On("GetMessages", mock.Anything,
		mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED}},
		mongostore.MessageOptions{Limit: int64(2)}).Return(messages, nil).Once()

	mockMessageClient.On("SendMessage", mock.Anything, "+905551234567", "Test message 1").
		Return(&messageclient.MessageResponse{MessageID: msgID1.Hex()}, nil).Once()
	mockMessageClient.On("SendMessage", mock.Anything, "+905559876543", "Test message 2").
		Return(&messageclient.MessageResponse{MessageID: msgID2.Hex()}, nil).Once()

	mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID1, mongostore.STATUS_SENT, mock.Anything).
		Return(nil).Once()
	mockMongoStore.On("UpdateMessageStatus", mock.Anything, msgID2, mongostore.STATUS_SENT, mock.Anything).
		Return(nil).Once()

	mockRedisStore.On("CacheMessageID", mock.Anything, msgID1.Hex()).
		Return(nil).Once()
	mockRedisStore.On("CacheMessageID", mock.Anything, msgID2.Hex()).
		Return(nil).Once()

	worker.process()

	mockMongoStore.AssertExpectations(t)
	mockMessageClient.AssertExpectations(t)
	mockRedisStore.AssertExpectations(t)
}
