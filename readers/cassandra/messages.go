// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

var errReadMessages = errors.New("failed to read messages from cassandra database")

const (
	format      = "format"
	defKeyspace = "messages"
)

var _ readers.MessageRepository = (*cassandraRepository)(nil)

type cassandraRepository struct {
	session *gocql.Session
}

// New instantiates Cassandra message repository.
func New(session *gocql.Session) readers.MessageRepository {
	return cassandraRepository{
		session: session,
	}
}

func (cr cassandraRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	keyspace, ok := query[format]
	if !ok {
		keyspace = defKeyspace
	}
	// Remove format filter and format the rest properly.
	delete(query, format)

	names := []string{}
	vals := []interface{}{chanID}
	for name, val := range query {
		names = append(names, name)
		vals = append(vals, val)
	}
	vals = append(vals, offset+limit)

	selectCQL := buildSelectQuery(keyspace, chanID, offset, limit, names)
	countCQL := buildCountQuery(keyspace, chanID, names)

	iter := cr.session.Query(selectCQL, vals...).Iter()
	defer iter.Close()
	scanner := iter.Scanner()

	// skip first OFFSET rows
	for i := uint64(0); i < offset; i++ {
		if !scanner.Next() {
			break
		}
	}

	page := readers.MessagesPage{
		Offset:   offset,
		Limit:    limit,
		Messages: []interface{}{},
	}

	switch keyspace {
	case defKeyspace:
		for scanner.Next() {
			var msg senml.Message
			err := scanner.Scan(&msg.Channel, &msg.Subtopic, &msg.Publisher, &msg.Protocol,
				&msg.Name, &msg.Unit, &msg.Value, &msg.StringValue, &msg.BoolValue,
				&msg.DataValue, &msg.Sum, &msg.Time, &msg.UpdateTime)
			if err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}
			page.Messages = append(page.Messages, msg)
		}
	default:
		for scanner.Next() {
			msg := map[string]interface{}{}
			err := scanner.Scan(&msg)
			if err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}
			page.Messages = append(page.Messages, parseFlat(msg))
		}
	}

	if err := cr.session.Query(countCQL, vals[:len(vals)-1]...).Scan(&page.Total); err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return page, nil
}

func buildSelectQuery(keyspace, chanID string, offset, limit uint64, names []string) string {
	var condCQL string
	cql := fmt.Sprintf(`SELECT channel, subtopic, publisher, protocol, name, unit,
	        value, string_value, bool_value, data_value, sum, time,
			update_time FROM %s WHERE channel = ? %s LIMIT ?
			ALLOW FILTERING`, keyspace, "%s")
	for _, name := range names {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condCQL = fmt.Sprintf(`%s AND %s = ?`, condCQL, name)
		}
	}

	return fmt.Sprintf(cql, condCQL)
}

func buildCountQuery(format, chanID string, names []string) string {
	var condCQL string
	cql := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE channel = ? %s ALLOW FILTERING`, format, "%s")

	for _, name := range names {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condCQL = fmt.Sprintf(`%s AND %s = ?`, condCQL, name)
		}
	}

	return fmt.Sprintf(cql, condCQL)
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
