// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	writer "github.com/mainflux/mainflux/consumers/writers/influxdb"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const valueFields = 5

var (
	port        string
	testLog, _  = log.New(os.Stdout, log.Info.String())
	streamsSize = 250
	selectMsgs  = "SELECT * FROM test..messages"
	dropMsgs    = "DROP SERIES FROM messages"
	writeAPI    api.WriteAPI
	queryAPI    api.QueryAPI
	deleteAPI   api.DeleteAPI
	authToken   = "secret"
	org         = "mainflux"
	bucket      = "messages"
	subtopic    = "topic"
)

var (
	v       float64 = 5
	stringV         = "value"
	boolV           = true
	dataV           = "base64"
	sum     float64 = 42
)

// This is utility function to query the database.
func queryDB(q string) ([][]interface{}, error) {
	response, err := queryAPI.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	defer response.Close()
	// if response. != nil {
	// 	return nil, response.Error()
	// }
	// if len(response.Results[0].Series) == 0 {
	// 	return nil, nil
	// }
	// There is only one query, so only one result and
	// all data are stored in the same series.
	for response.Next() {
		response.Record()

	}
	return nil, nil
}

func TestSave(t *testing.T) {
	repo := writer.New(writeAPI)

	cases := []struct {
		desc         string
		msgsNum      int
		expectedSize int
	}{
		{
			desc:         "save a single message",
			msgsNum:      1,
			expectedSize: 1,
		},
		{
			desc:         "save a batch of messages",
			msgsNum:      streamsSize,
			expectedSize: streamsSize,
		},
	}

	for _, tc := range cases {
		// Clean previously saved messages.
		// _, err := queryDB(dropMsgs)
		err := deleteAPI.DeleteWithName(context.Background(), org, bucket, time.Time{}, time.Now(), "_measurement=messages")
		require.Nil(t, err, fmt.Sprintf("Cleaning data from InfluxDB expected to succeed: %s.\n", err))

		now := time.Now().UnixNano()
		msg := senml.Message{
			Channel:    "45",
			Publisher:  "2580",
			Protocol:   "http",
			Name:       "test name",
			Unit:       "km",
			UpdateTime: 5456565466,
		}
		var msgs []senml.Message

		for i := 0; i < tc.msgsNum; i++ {
			// Mix possible values as well as value sum.
			count := i % valueFields
			switch count {
			case 0:
				msg.Subtopic = subtopic
				msg.Value = &v
			case 1:
				msg.BoolValue = &boolV
			case 2:
				msg.StringValue = &stringV
			case 3:
				msg.DataValue = &dataV
			case 4:
				msg.Sum = &sum
			}

			msg.Time = float64(now)/float64(1e9) + float64(i)
			msgs = append(msgs, msg)
		}

		err = repo.Consume(msgs)
		assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed: %s.\n", err))

		// row, err := queryDB(selectMsgs)
		// assert.Nil(t, err, fmt.Sprintf("Querying InfluxDB to retrieve data expected to succeed: %s.\n", err))

		// count := len(row)
		// assert.Equal(t, tc.expectedSize, count, fmt.Sprintf("Expected to have %d messages saved, found %d instead.\n", tc.expectedSize, count))
	}
}
