package mockmongostore

import (
	"context"
	"fmt"
	"github.com/mkaykisiz/sender"
	"github.com/stretchr/testify/mock"
	"time"

	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// compile-time proof of mongo store interface implementation
var _ mongostore.Store = (*Store)(nil)

// Store represents mock mongo store
type Store struct {
	mock.Mock
}

// NewStore returns mock mongo store
func NewStore() *Store {
	return &Store{}
}

// ReadPaymentTransaction mocks read payment transaction
func (s *Store) GetMessages(ctx context.Context, f mongostore.MessageFilter, o mongostore.MessageOptions) ([]sender.MessageTransaction, error) {
	args := s.Called(ctx, f, o)
	return args.Get(0).([]sender.MessageTransaction), args.Error(1)
}

// UpdateMessageStatus mocks update message status
func (s *Store) UpdateMessageStatus(ctx context.Context, id primitive.ObjectID, status string, sentAt *time.Time) error {
	fmt.Printf("Mock called with: id=%v, status=%v, sentAt=%v\n", id, status, sentAt)
	args := s.Called(ctx, id, status, sentAt)
	return args.Error(0)
}

// Count mocks count
func (s *Store) Count(ctx context.Context, f mongostore.MessageFilter) (int64, error) {
	args := s.Called(ctx, f)
	return args.Get(0).(int64), args.Error(1)
}

// Insert mocks insert
func (s *Store) Insert(ctx context.Context, mt sender.MessageTransaction) error {
	args := s.Called(ctx, mt)
	return args.Error(0)
}

// InsertMany mocks insert many
func (s *Store) InsertMany(ctx context.Context, mts []sender.MessageTransaction) error {
	args := s.Called(ctx, mts)
	return args.Error(0)
}

// Close mocks to close method
func (s *Store) Close() error {
	args := s.Called()

	return args.Error(0)
}
