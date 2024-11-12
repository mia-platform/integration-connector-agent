// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewWriter(t *testing.T) {
	t.Parallel()

	valid := validateFunc(func(context.Context, *mongo.Client) error { return nil })
	invalid := validateFunc(func(context.Context, *mongo.Client) error { return errors.New("invalid") })

	tests := map[string]struct {
		configuration  *Config
		expectedWriter *Writer[entities.PipelineEvent]
		validateFunc   validateFunc
		expectedError  bool
	}{
		"invalid connection string return error": {
			configuration: &Config{
				URL:        "invalid://uri/for/mongo",
				Database:   "foo",
				Collection: "bar",
			},
			validateFunc:  valid,
			expectedError: true,
		},
		"cannot receive ping from url return error": {
			configuration: &Config{
				URL:        "mongodb://localhost:27018/baz?connectTimeoutMS=200",
				Collection: "bar",
			},
			validateFunc:  invalid,
			expectedError: true,
		},
		"valid uri return writer": {
			configuration: &Config{
				URL:        "mongodb://localhost:27017/baz?connectTimeoutMS=200",
				Collection: "bar",
			},
			validateFunc: valid,
			expectedWriter: &Writer[entities.PipelineEvent]{
				collection: "bar",
				database:   "baz",
			},
		},
		"valid uri withtout database return writer": {
			configuration: &Config{
				URL:        "mongodb://localhost:27017/?connectTimeoutMS=200",
				Database:   "baz",
				Collection: "bar",
			},
			validateFunc: valid,
			expectedWriter: &Writer[entities.PipelineEvent]{
				collection: "bar",
				database:   "baz",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			writer, err := newMongoDBWriter[entities.PipelineEvent](ctx, test.configuration, test.validateFunc)
			switch test.expectedError {
			case false:
				assert.NoError(t, err)
				require.NotNil(t, writer)
				mongoWriter, ok := writer.(*Writer[entities.PipelineEvent])
				require.True(t, ok)
				assert.NotNil(t, mongoWriter.client)
				assert.Equal(t, test.expectedWriter.collection, mongoWriter.collection)
				assert.Equal(t, test.expectedWriter.database, mongoWriter.database)
			case true:
				assert.ErrorContains(t, err, ErrMongoInitialization.Error())
				assert.Nil(t, writer)
			}
		})
	}
}

func TestUpsert(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		data        entities.PipelineEvent
		responses   primitive.D
		expectedErr string
	}{
		"upserting element": {
			data:      getEvent(t, nil),
			responses: mtest.CreateSuccessResponse(bson.E{Key: "upserted", Value: []any{bson.D{}}}),
		},
		"update element": {
			data:      getEvent(t, nil),
			responses: mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: 1}),
		},
		"error if event contains reserved data": {
			data:        getEvent(t, map[string]any{idField: "reserved"}),
			expectedErr: "event data contains reserved field eventId",
		},
		"error without change": {
			data:        getEvent(t, nil),
			responses:   mtest.CreateSuccessResponse(bson.E{}),
			expectedErr: "error upserting data: 0 documents upserted",
		},
	}

	for testName, test := range tests {
		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Run(testName, func(mt *mtest.T) {
			writer := &Writer[entities.PipelineEvent]{
				client:     mt.Client,
				collection: mt.Coll.Name(),
				database:   mt.DB.Name(),
				idField:    idField,
			}

			mt.AddMockResponses(test.responses)

			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			err := writer.Write(ctx, test.data)
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		data        entities.PipelineEvent
		responses   primitive.D
		expectedErr string
	}{
		"delete element": {
			data:      getEvent(t, nil),
			responses: mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}),
		},
		"error if event without id": {
			data:        &entities.Event{},
			expectedErr: "event id is empty",
		},
		"error without change": {
			data:        getEvent(t, nil),
			responses:   mtest.CreateSuccessResponse(bson.E{}),
			expectedErr: "error deleting data: 0 documents deleted",
		},
	}

	for testName, test := range tests {
		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Run(testName, func(mt *mtest.T) {
			writer := &Writer[entities.PipelineEvent]{
				client:     mt.Client,
				collection: mt.Coll.Name(),
				database:   mt.DB.Name(),
			}

			mt.AddMockResponses(test.responses)

			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			err := writer.Delete(ctx, test.data)
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func getEvent(t *testing.T, data map[string]any) entities.PipelineEvent {
	t.Helper()

	event := &entities.Event{
		ID: "12345",
	}

	if data != nil {
		event.WithData(data)
	}

	return event
}

func TestOutputModel(t *testing.T) {
	t.Parallel()

	outputModel := map[string]any{}
	config := &Config{
		URL:         config.SecretSource("mongodb://localhost:27017/?connectTimeoutMS=200"),
		Database:    "foo",
		Collection:  "bar",
		OutputEvent: outputModel,
	}
	valid := validateFunc(func(context.Context, *mongo.Client) error { return nil })

	writer, err := newMongoDBWriter[entities.PipelineEvent](context.Background(), config, valid)
	require.NoError(t, err)
	require.Equal(t, outputModel, writer.OutputModel())
}
