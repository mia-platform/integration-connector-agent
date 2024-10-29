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
	"path/filepath"
	"testing"
	"time"

	"github.com/mia-platform/data-connector-agent/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type fakeWriterData struct {
	id string
}

func (f fakeWriterData) ID() string {
	return f.id
}

func TestNewWriter(t *testing.T) {
	t.Parallel()

	valid := validateFunc(func(context.Context, *mongo.Client) error { return nil })
	invalid := validateFunc(func(context.Context, *mongo.Client) error { return errors.New("invalid") })

	tests := map[string]struct {
		configuration  Config
		expectedWriter *Writer[fakeWriterData]
		validateFunc   validateFunc
		expectedError  bool
	}{
		"invalid connection string return error": {
			configuration: Config{
				URI: utils.SecretSource{
					FromFile: filepath.Join("testdata", "invaliduri"),
				},
				Database:   "foo",
				Collection: "bar",
			},
			validateFunc:  valid,
			expectedError: true,
		},
		"cannot receive ping from url return error": {
			configuration: Config{
				URI: utils.SecretSource{
					FromFile: filepath.Join("testdata", "missingserver"),
				},
				Collection: "bar",
			},
			validateFunc:  invalid,
			expectedError: true,
		},
		"valid uri return writer": {
			configuration: Config{
				URI: utils.SecretSource{
					FromFile: filepath.Join("testdata", "validuri"),
				},
				Collection: "bar",
			},
			validateFunc: valid,
			expectedWriter: &Writer[fakeWriterData]{
				collection: "bar",
				database:   "baz",
			},
		},
		"valid uri withtout database return writer": {
			configuration: Config{
				URI: utils.SecretSource{
					FromFile: filepath.Join("testdata", "validuri-without-db"),
				},
				Database:   "baz",
				Collection: "bar",
			},
			validateFunc: valid,
			expectedWriter: &Writer[fakeWriterData]{
				collection: "bar",
				database:   "baz",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			writer, err := newMongoDBWriter[fakeWriterData](ctx, test.configuration, test.validateFunc)
			switch test.expectedError {
			case false:
				assert.NoError(t, err)
				require.NotNil(t, writer)
				mongoWriter, ok := writer.(*Writer[fakeWriterData])
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
		data        fakeWriterData
		responses   primitive.D
		expectedErr bool
	}{
		"no error": {
			data: fakeWriterData{id: "12345"},
			responses: bson.D{
				{Key: "ok", Value: 1},
				{Key: "value", Value: bson.D{}},
			},
		},
		"error": {
			data:        fakeWriterData{},
			expectedErr: true,
		},
	}

	for testName, test := range tests {
		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Run(testName, func(mt *mtest.T) {
			writer := &Writer[fakeWriterData]{
				client:     mt.Client,
				collection: mt.Coll.Name(),
				database:   mt.DB.Name(),
			}

			mt.AddMockResponses(test.responses)

			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			err := writer.Write(ctx, test.data)
			switch test.expectedErr {
			case true:
				assert.Error(mt, err)
			case false:
				assert.NoError(mt, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		data        fakeWriterData
		responses   primitive.D
		expectedErr bool
	}{
		"no error": {
			data: fakeWriterData{id: "12345"},
			responses: bson.D{
				{Key: "ok", Value: 1},
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: "12345"},
				}},
			},
		},
		"error": {
			data:        fakeWriterData{},
			expectedErr: true,
		},
	}

	for testName, test := range tests {
		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Run(testName, func(mt *mtest.T) {
			writer := &Writer[fakeWriterData]{
				client:     mt.Client,
				collection: mt.Coll.Name(),
				database:   mt.DB.Name(),
			}

			mt.AddMockResponses(test.responses)

			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()

			err := writer.Delete(ctx, test.data)
			switch test.expectedErr {
			case true:
				assert.Error(mt, err)
			case false:
				assert.NoError(mt, err)
			}
		})
	}
}
