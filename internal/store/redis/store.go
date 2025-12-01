package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	envvars "github.com/mkaykisiz/sender/configs/env-vars"
)

const (
	messageExpiration = 1 * time.Hour
)

// Store defines behaviors of redis store
type Store interface {
	CacheMessageID(ctx context.Context, id string) error
	Close() error
}

// Store represents redis store
type store struct {
	address            string
	password           string
	db                 int
	dialTimeout        time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	poolSize           int
	minIdleConnections int
	maxConnectionAge   time.Duration
	idleTimeout        time.Duration
	c                  *redis.Client
	mtx                sync.Mutex
}

// NewStore creates and returns redis store
func NewStore(r envvars.Redis) (Store, error) {
	s := &store{
		address:            r.Address,
		password:           r.Password,
		db:                 r.DB,
		dialTimeout:        r.DialTimeout,
		readTimeout:        r.ReadTimeout,
		writeTimeout:       r.WriteTimeout,
		poolSize:           r.PoolSize,
		minIdleConnections: r.MinIdleConnections,
		maxConnectionAge:   r.MaxConnectionAge,
		idleTimeout:        r.IdleTimeout,
	}

	opts := &redis.Options{
		Addr:         s.address,
		Password:     s.password,
		DB:           s.db,
		DialTimeout:  s.dialTimeout,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		PoolSize:     s.poolSize,
		MinIdleConns: s.minIdleConnections,
		MaxConnAge:   s.maxConnectionAge,
		IdleTimeout:  s.idleTimeout,
	}

	c := redis.NewClient(opts)

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("pinging failed, %s", err.Error())
	}

	s.c = c

	return s, nil
}

func (s *store) CacheMessageID(ctx context.Context, id string) error {
	data := time.Now()
	if err := s.c.Set(ctx, id, data, messageExpiration).Err(); err != nil {
		return fmt.Errorf("setting msg failed, %s", err.Error())
	}

	return nil
}

// Close closes underlying redis client
func (s *store) Close() error {
	return s.c.Close()
}
