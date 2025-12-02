package mongostore

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	STATUS_PENDING = "pending"
	STATUS_SENT    = "sent"
	STATUS_FAILED  = "failed"
	STATUS_INVALID = "invalid"
)

type MessageFilter struct {
	Status []string
}

func (f MessageFilter) ToFilter(baseFilter bson.M) bson.M {
	if len(f.Status) > 0 {
		baseFilter["status"] = bson.M{"$in": f.Status}
	}

	return baseFilter
}

type MessageOptions struct {
	Limit int64
}

func (f MessageOptions) ToOptions() *options.FindOptions {
	options := options.Find()

	if f.Limit > 0 {
		options.SetLimit(f.Limit)
	}

	return options
}
