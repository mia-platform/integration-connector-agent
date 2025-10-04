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
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoUpsertUnit(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("mongo upsert operations", func(mt *mtest.T) {
		ctx := mt.Context()

		w := &Writer[*entities.Event]{
			client:        mt.Client,
			database:      mt.DB.Name(),
			collection:    mt.Coll.Name(),
			upsertIDField: "_eventId",
			insertOnly:    false,
		}

		primaryKey := entities.PkFields{{Key: "test", Value: "123"}}

		t.Run("create", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "upserted", Value: []any{bson.D{}}}))

			e := getTestEventUnit(t, primaryKey, map[string]any{"foo": "bar", "key": "123"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("update", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: 1}))

			e := getTestEventUnit(t, primaryKey, map[string]any{"foo": "taz", "key": "123", "another": "field"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("delete", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

			e := getTestEventUnit(t, primaryKey, nil, entities.Delete)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})
	})
}

func TestMongoUpsertWithMultiplePrimaryKeysUnit(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("mongo upsert with multiple primary keys", func(mt *mtest.T) {
		ctx := mt.Context()

		w := &Writer[*entities.Event]{
			client:        mt.Client,
			database:      mt.DB.Name(),
			collection:    mt.Coll.Name(),
			upsertIDField: "_eventId",
			insertOnly:    false,
		}

		primaryKeys := entities.PkFields{
			{Key: "id1", Value: "123"},
			{Key: "id2", Value: "456"},
		}

		t.Run("create", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "upserted", Value: []any{bson.D{}}}))

			e := getTestEventUnit(t, primaryKeys, map[string]any{"foo": "bar", "key": "123"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("update", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: 1}))

			e := getTestEventUnit(t, primaryKeys, map[string]any{"foo": "taz", "key": "123", "another": "field"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("delete", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

			e := getTestEventUnit(t, primaryKeys, nil, entities.Delete)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})
	})
}

func TestMongoOnlyInsertUnit(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("mongo insert only operations", func(mt *mtest.T) {
		ctx := mt.Context()

		w := &Writer[*entities.Event]{
			client:        mt.Client,
			database:      mt.DB.Name(),
			collection:    mt.Coll.Name(),
			upsertIDField: "_eventId",
			insertOnly:    true,
		}

		primaryKey := entities.PkFields{{Key: "key", Value: "234"}}

		t.Run("insert new data - 1", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse())

			e := getTestEventUnit(t, primaryKey, map[string]any{"foo": "bar", "key": "234", "type": "created"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("insert new data with existing id already saved", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse())

			e := getTestEventUnit(t, primaryKey, map[string]any{"foo": "taz", "key": "234", "another": "field", "type": "updated"}, entities.Write)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})

		t.Run("insert new deletion data", func(t *testing.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse())

			e := getTestEventUnit(t, primaryKey, map[string]any{"foo": "taz", "key": "234", "another": "field", "type": "deleted"}, entities.Delete)
			err := w.WriteData(ctx, e)
			require.NoError(t, err)
		})
	})
}

func TestMongoWriterCreationUnit(t *testing.T) {
	t.Run("writer creation with mock validation", func(t *testing.T) {
		ctx := t.Context()

		// Test successful writer creation with mocked validation
		w, err := newMongoDBWriter[*entities.Event](ctx, &Config{
			URL:        config.SecretSource("mongodb://localhost:27017/testdb"),
			Database:   "testdb",
			Collection: "testcoll",
		}, func(context.Context, *mongo.Client) error {
			return nil // Mock successful validation
		})

		require.NoError(t, err)
		require.NotNil(t, w)

		mongoWriter, ok := w.(*Writer[*entities.Event])
		require.True(t, ok)
		require.Equal(t, "testdb", mongoWriter.database)
		require.Equal(t, "testcoll", mongoWriter.collection)
		require.Equal(t, "_eventId", mongoWriter.upsertIDField)
		require.False(t, mongoWriter.insertOnly)
	})

	t.Run("writer creation with insert only mode", func(t *testing.T) {
		ctx := t.Context()

		w, err := newMongoDBWriter[*entities.Event](ctx, &Config{
			URL:        config.SecretSource("mongodb://localhost:27017/testdb"),
			Database:   "testdb",
			Collection: "testcoll",
			InsertOnly: true,
		}, func(context.Context, *mongo.Client) error {
			return nil // Mock successful validation
		})

		require.NoError(t, err)
		require.NotNil(t, w)

		mongoWriter, ok := w.(*Writer[*entities.Event])
		require.True(t, ok)
		require.True(t, mongoWriter.insertOnly)
	})
}

func getTestEventUnit(t *testing.T, pks entities.PkFields, data map[string]any, operation entities.Operation) *entities.Event {
	t.Helper()

	e := &entities.Event{
		PrimaryKeys:   pks,
		OperationType: operation,
	}

	if data != nil {
		d, err := json.Marshal(data)
		require.NoError(t, err)
		e.WithData(d)
	}

	return e
}
