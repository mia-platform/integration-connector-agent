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

package crudservice

import (
	"context"
	"errors"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/mia-platform/go-crud-service-client"
	"github.com/mia-platform/go-crud-service-client/testhelper/mock"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	t.Run("delete event with primary keys", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			pkFieldPrefix: "pk",
			c: &mock.CRUD[any]{
				DeleteManyResult: 1,
				DeleteManyAssertionFunc: func(_ context.Context, options crud.Options) {
					require.Equal(t, map[string]any{"pk.key1": "12345", "pk.key2": "12345"}, options.Filter.MongoQuery)
				},
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "12345"},
			},
		}
		err := client.Delete(context.Background(), event)
		require.NoError(t, err)
	})

	t.Run("fails on delete error", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			c: &mock.CRUD[any]{
				DeleteManyError: errors.New("some error from crud"),
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "12345"},
			},
		}
		err := client.Delete(context.Background(), event)
		require.Error(t, err)
	})
}

func TestUpsert(t *testing.T) {
	t.Run("successfully update event with primary keys", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			pkFieldPrefix: "pk",
			c: &mock.CRUD[any]{
				UpsertOneAssertionFunc: func(_ context.Context, body crud.UpsertBody, options crud.Options) {
					require.Equal(t, map[string]any{"pk.key1": "12345", "pk.key2": "98765"}, options.Filter.MongoQuery)
					require.Equal(t, map[string]any{
						"data": "some data",
						"pk": map[string]string{
							"key1": "12345",
							"key2": "98765",
						},
					}, body.Set)
				},
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`{"data": "some data"}`),
		}
		err := client.Upsert(context.Background(), event)
		require.NoError(t, err)
	})

	t.Run("failure on json serialization error", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			c: &mock.CRUD[any]{
				UpsertOneAssertionFunc: func(_ context.Context, _ crud.UpsertBody, _ crud.Options) {
					t.Fatalf("should not reach this point, expected json serialization error")
				},
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`This ain't valid JSON`),
		}
		err := client.Upsert(context.Background(), event)
		require.ErrorContains(t, err, "invalid character")
	})

	t.Run("failure on client error", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			pkFieldPrefix: "pk",
			c: &mock.CRUD[any]{
				UpsertOneAssertionFunc: func(_ context.Context, body crud.UpsertBody, options crud.Options) {
					require.Equal(t, map[string]any{"pk.key1": "12345", "pk.key2": "98765"}, options.Filter.MongoQuery)
					require.Equal(t, map[string]any{
						"data": "some data",
						"pk": map[string]string{
							"key1": "12345",
							"key2": "98765",
						},
					}, body.Set)
				},
				UpsertOneError: errors.New("some error from crud"),
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`{"data": "some data"}`),
		}
		err := client.Upsert(context.Background(), event)
		require.ErrorContains(t, err, "some error from crud")
	})
}

func TestInsert(t *testing.T) {
	t.Run("successfully insert event with primary keys", func(t *testing.T) {
		var invoked bool
		client := &client[entities.PipelineEvent]{
			pkFieldPrefix: "pk",
			c: &mock.CRUD[any]{
				CreateAssertionFunc: func(_ context.Context, body any, options crud.Options) {
					invoked = true
					require.Empty(t, options.Filter.MongoQuery)
					require.Equal(t, map[string]any{
						"data": "some data",
						"pk": map[string]string{
							"key1": "12345",
							"key2": "98765",
						},
					}, body)
				},
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`{"data": "some data"}`),
		}
		err := client.Insert(context.Background(), event)
		require.NoError(t, err)
		require.True(t, invoked, "Create should have been called")
	})

	t.Run("failure on json serialization error", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			c: &mock.CRUD[any]{
				CreateAssertionFunc: func(_ context.Context, _ any, _ crud.Options) {
					t.Fatalf("should not reach this point, expected json serialization error")
				},
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`This ain't valid JSON`),
		}
		err := client.Insert(context.Background(), event)
		require.ErrorContains(t, err, "invalid character")
	})

	t.Run("failure on client error", func(t *testing.T) {
		client := &client[entities.PipelineEvent]{
			pkFieldPrefix: "pk",
			c: &mock.CRUD[any]{
				CreateAssertionFunc: func(_ context.Context, body any, options crud.Options) {
					require.Empty(t, options.Filter.MongoQuery)
					require.Equal(t, map[string]any{
						"data": "some data",
						"pk": map[string]string{
							"key1": "12345",
							"key2": "98765",
						},
					}, body)
				},
				CreateError: errors.New("some error from crud"),
			},
		}

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "key1", Value: "12345"},
				{Key: "key2", Value: "98765"},
			},
			OriginalRaw: []byte(`{"data": "some data"}`),
		}
		err := client.Insert(context.Background(), event)
		require.ErrorContains(t, err, "some error from crud")
	})
}
