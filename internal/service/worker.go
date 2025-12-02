package service

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mkaykisiz/sender"
	"github.com/mkaykisiz/sender/internal/client/messageclient"
	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	redisstore "github.com/mkaykisiz/sender/internal/store/redis"
)

const MaxMessageLength = 1000

type Worker struct {
	sender  messageclient.MessageClient
	ms      mongostore.Store
	rs      redisstore.Store
	l       log.Logger
	ticker  *time.Ticker
	done    chan bool
	running bool
	limit   int64
	mu      sync.Mutex
}

// NewWorker creates and returns worker
func NewWorker(sender messageclient.MessageClient, ms mongostore.Store, rs redisstore.Store, l log.Logger, limit int64) *Worker {
	return &Worker{
		sender:  sender,
		ms:      ms,
		rs:      rs,
		l:       l,
		done:    make(chan bool),
		running: false,
		limit:   limit,
	}
}

func (w *Worker) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return
	}
	w.running = true
	w.ticker = time.NewTicker(2 * time.Minute)
	w.done = make(chan bool)

	go func() {
		w.process() // Run first time
		for {
			select {
			case <-w.done:
				return
			case <-w.ticker.C:
				w.process() // Run every x minutes
			}
		}
	}()
	w.logWithLogger(nil, map[string]interface{}{
		"method": "Start",
		"msg":    "started",
	})
}

func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}
	w.running = false
	w.ticker.Stop()
	close(w.done)
	w.logWithLogger(nil, map[string]interface{}{
		"method": "Stop",
		"msg":    "stopped",
	})
}

func (w *Worker) process() {
	w.logWithLogger(nil, map[string]interface{}{
		"method": "process",
		"msg":    "processing",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messageFilter := mongostore.MessageFilter{
		Status: []string{mongostore.STATUS_PENDING, mongostore.STATUS_FAILED},
	}
	messageOptions := mongostore.MessageOptions{
		Limit: w.limit,
	}
	messages, err := w.ms.GetMessages(ctx, messageFilter, messageOptions)
	if err != nil {
		w.logWithLogger(err, map[string]interface{}{
			"method": "process",
			"msg":    "error getting messages",
		})
		return
	}

	if len(messages) == 0 {
		w.logWithLogger(nil, map[string]interface{}{
			"method": "process",
			"msg":    "no unsent messages found",
		})
		return
	}

	// Send messages concurrently
	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg sender.MessageTransaction) {
			defer wg.Done()

			if !msg.IsValid() {
				// Update status to INVALID
				w.logWithLogger(nil, map[string]interface{}{
					"method": "process",
					"msg":    "message is invalid",
					"id":     msg.ID,
				})
				err = w.ms.UpdateMessageStatus(ctx, msg.ID, mongostore.STATUS_INVALID, nil)
				if err != nil {
					w.logWithLogger(err, map[string]interface{}{
						"method": "process",
						"msg":    "error updating message status to INVALID",
						"id":     msg.ID,
					})
				}
				return
			}

			res, err := w.sender.SendMessage(ctx, msg.Recipient, msg.Content)
			if err != nil {
				w.logWithLogger(err, map[string]interface{}{
					"method": "process",
					"msg":    "error sending message, trying to update status to FAILED",
					"id":     msg.ID,
				})

				// Retry 3 times to update status to FAILED
				for i := 0; i < 3; i++ {
					err = w.ms.UpdateMessageStatus(ctx, msg.ID, mongostore.STATUS_FAILED, nil)
					if err == nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				if err != nil {
					w.logWithLogger(err, map[string]interface{}{
						"method": "process",
						"msg":    "error updating messages",
						"id":     msg.ID,
					})
				}
				return
			}

			// Retry updating status to SENT
			for i := 0; i < 3; i++ {
				now := time.Now()
				err = w.ms.UpdateMessageStatus(ctx, msg.ID, mongostore.STATUS_SENT, &now)
				if err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if err != nil {
				w.logWithLogger(err, map[string]interface{}{
					"method": "process",
					"msg":    "error updating messages",
					"id":     msg.ID,
				})
				return
			}
			// Cache message id
			err = w.rs.CacheMessageID(ctx, res.MessageID)
			if err != nil {
				w.logWithLogger(err, map[string]interface{}{
					"method": "process",
					"msg":    "error caching message id",
					"id":     msg.ID,
				})
				return
			}

			w.logWithLogger(nil, map[string]interface{}{
				"method": "process",
				"msg":    "message sent successfully",
				"id":     msg.ID,
			})
		}(msg)
	}
	wg.Wait()
}

func (w *Worker) logWithLogger(err error, additionalParams map[string]interface{}) {
	logParams := make([]interface{}, 0, 2+len(additionalParams)*2)

	for k, v := range additionalParams {
		logParams = append(logParams, k, v)
	}

	if err != nil {
		logParams = append(logParams, "error", err.Error())
		_ = level.Error(w.l).Log(logParams...)
	} else {
		_ = level.Info(w.l).Log(logParams...)
	}
}
