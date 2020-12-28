// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	format        = "format"
	defCollection = "messages"
)

var errReadMessages = errors.New("failed to read messages from mongodb database")

var _ readers.MessageRepository = (*mongoRepository)(nil)

type mongoRepository struct {
	db *mongo.Database
}

// New returns new MongoDB reader.
func New(db *mongo.Database) readers.MessageRepository {
	return mongoRepository{
		db: db,
	}
}

func (repo mongoRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	format, ok := query[format]
	if !ok {
		format = defCollection
	}
	col := repo.db.Collection(format)
	sortMap := map[string]interface{}{
		"time": -1,
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	filter := fmtCondition(chanID, query)
	cursor, err := col.Find(context.Background(), filter, options.Find().SetSort(sortMap).SetLimit(int64(limit)).SetSkip(int64(offset)))
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer cursor.Close(context.Background())

	messages := []interface{}{}
	// var m interface{}
	switch format {
	case defCollection:
		for cursor.Next(context.Background()) {
			var m senml.Message
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}

			messages = append(messages, m)
		}
	default:
		for cursor.Next(context.Background()) {
			var m map[string]interface{}
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}

			messages = append(messages, parseFlat(m))
		}
	}

	total, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if total < 0 {
		return readers.MessagesPage{}, nil
	}

	return readers.MessagesPage{
		Total:    uint64(total),
		Offset:   offset,
		Limit:    limit,
		Messages: messages,
	}, nil
}

func fmtCondition(chanID string, query map[string]string) *bson.D {
	filter := bson.D{
		bson.E{
			Key:   "channel",
			Value: chanID,
		},
	}
	for name, value := range query {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			filter = append(filter, bson.E{Key: name, Value: value})
		}
	}

	return &filter
}

func parseFlat(flat interface{}) interface{} {
	msg := make(map[string]interface{})
	switch v := flat.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			keys := strings.Split(key, "/")
			n := len(keys)
			if n == 1 {
				msg[key] = value
				continue
			}
			current := msg
			for i, k := range keys {
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				if i == n-1 {
					current[k] = value
					break
				}
				current = current[k].(map[string]interface{})
			}
		}
	}
	return msg
}
