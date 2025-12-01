package mockredisstore

import (
	"context"
	"github.com/stretchr/testify/mock"

	redisstore "github.com/mkaykisiz/sender/internal/store/redis"
)

// compile-time proof of redis store interface implementation
var _ redisstore.Store = (*Store)(nil)

// Store represents mock redis store
type Store struct {
	mock.Mock
}

// NewStore returns mock redis store
func NewStore() *Store {
	return &Store{}
}

// CacheMessageID mocks cache message id method
func (s *Store) CacheMessageID(ctx context.Context, id string) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

// Close mocks to close method
func (s *Store) Close() error {
	args := s.Called()

	return args.Error(0)
}
