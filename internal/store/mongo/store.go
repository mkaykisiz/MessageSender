package mongostore

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mkaykisiz/sender"
	envvars "github.com/mkaykisiz/sender/configs/env-vars"
	"golang.org/x/net/context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MessageCollectionName = "message"
)

// Store defines behaviors of mongo store
type Store interface {
	Close() error
	GetMessages(ctx context.Context, f MessageFilter, o MessageOptions) (mts []sender.MessageTransaction, err error)
	UpdateMessageStatus(ctx context.Context, id primitive.ObjectID, status string, sentAt *time.Time) error
	Count(ctx context.Context, f MessageFilter) (int64, error)
	InsertMany(ctx context.Context, mts []sender.MessageTransaction) error
}

// store represents mongo store
type store struct {
	uri               string
	database          string
	connectTimeout    time.Duration
	pingTimeout       time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	disconnectTimeout time.Duration
	c                 *mongo.Client
	db                *mongo.Database
}

// NewStore creates and returns mongo store
func NewStore(m envvars.Mongo) (*store, error) {
	s := &store{
		uri:               m.URI,
		database:          m.Database,
		connectTimeout:    m.ConnectTimeout,
		readTimeout:       m.ReadTimeout,
		writeTimeout:      m.WriteTimeout,
		pingTimeout:       m.PingTimeout,
		disconnectTimeout: m.DisconnectTimeout,
	}

	cctx, ccf := context.WithTimeout(context.Background(), s.connectTimeout)
	defer ccf()

	opts := options.Client()
	opts.ApplyURI(s.uri)

	c, err := mongo.Connect(cctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connecting failed, %s", err.Error())
	}

	s.c = c

	pctx, pcf := context.WithTimeout(context.Background(), s.pingTimeout)
	defer pcf()

	if err := s.c.Ping(pctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("pinging failed, %s", err.Error())
	}

	s.db = c.Database(s.database)

	return s, nil
}

func (s *store) GetMessages(ctx context.Context, f MessageFilter, o MessageOptions) ([]sender.MessageTransaction, error) {
	ctx, cf := context.WithTimeout(ctx, s.readTimeout)
	defer cf()

	var messageTransactions []sender.MessageTransaction

	findOptions := o.ToOptions()
	findOptions.SetSort(bson.D{{"created_at", 1}})

	cursor, err := s.db.Collection(MessageCollectionName).Find(ctx, f.ToFilter(bson.M{}), findOptions)
	if err != nil {
		return messageTransactions, err
	}

	err = cursor.All(ctx, &messageTransactions)
	if err != nil {
		return messageTransactions, err
	}
	return messageTransactions, nil
}

func (s *store) UpdateMessageStatus(ctx context.Context, id primitive.ObjectID, status string, sentAt *time.Time) error {
	ctx, cf := context.WithTimeout(ctx, s.writeTimeout)
	defer cf()

	update := bson.M{"status": status}
	if sentAt != nil {
		update["sent_at"] = sentAt
	}

	_, err := s.db.Collection(MessageCollectionName).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	return nil
}

func (s *store) Count(ctx context.Context, f MessageFilter) (int64, error) {
	ctx, cf := context.WithTimeout(ctx, s.readTimeout)
	defer cf()

	count, err := s.db.Collection(MessageCollectionName).CountDocuments(ctx, f.ToFilter(bson.M{}))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *store) InsertMany(ctx context.Context, mts []sender.MessageTransaction) error {
	ctx, cf := context.WithTimeout(ctx, s.writeTimeout)
	defer cf()

	var documents []interface{}
	for _, mt := range mts {
		documents = append(documents, mt)
	}

	_, err := s.db.Collection(MessageCollectionName).InsertMany(ctx, documents)
	if err != nil {
		return err
	}
	return nil
}

// Close disconnects underlying mongo client
func (s *store) Close() error {
	ctx, cf := context.WithTimeout(context.Background(), s.disconnectTimeout)
	defer cf()

	return s.c.Disconnect(ctx)
}
