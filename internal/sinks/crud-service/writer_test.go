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
	"net/http"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/mia-platform/go-crud-service-client"
	gock "github.com/mia-platform/go-crud-service-client/testhelper/gock"
	"github.com/stretchr/testify/require"
)

func TestWriteData(t *testing.T) {
	t.Run("delete operation", func(t *testing.T) {
		t.Run("successful delete", func(t *testing.T) {
			filter := crud.Filter{
				MongoQuery: map[string]any{
					"_pk.key1": "12345",
					"_pk.key2": "98765",
				},
			}

			gock.NewGockScope(t, "http://example.com/crud", http.MethodDelete, "").
				AddMatcher(gock.CrudQueryMatcher(t, gock.Filter(filter))).
				Reply(200).BodyString("1")

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL: "http://example.com/crud",
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Delete,
			})
			require.NoError(t, err)
		})

		t.Run("delete failed with status code != 200", func(t *testing.T) {
			filter := crud.Filter{
				MongoQuery: map[string]any{
					"_pk.key1": "12345",
					"_pk.key2": "98765",
				},
			}

			gock.NewGockScope(t, "http://example.com/crud", http.MethodDelete, "").
				AddMatcher(gock.CrudQueryMatcher(t, gock.Filter(filter))).
				Reply(500).JSON(map[string]any{"error": "Internal Server Error"})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL: "http://example.com/crud",
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Delete,
			})
			require.Error(t, err)
		})

		t.Run("inserts event if insertOnly is set to true", func(t *testing.T) {
			gock.NewGockScope(t, "http://example.com/crud/", http.MethodPost, "").
				Reply(200).JSON(map[string]any{})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL:        "http://example.com/crud/",
					InsertOnly: true,
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Delete,
				OriginalRaw:   []byte(`{"data": "some data"}`),
			})
			require.NoError(t, err)
		})
	})

	t.Run("write operation", func(t *testing.T) {
		t.Run("successful upsert", func(t *testing.T) {
			filter := crud.Filter{
				MongoQuery: map[string]any{
					"_pk.key1": "12345",
					"_pk.key2": "98765",
				},
			}

			gock.NewGockScope(t, "http://example.com/crud/", http.MethodPost, "upsert-one").
				AddMatcher(gock.CrudQueryMatcher(t, gock.Filter(filter))).
				Reply(200).JSON(map[string]any{})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL: "http://example.com/crud/",
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Write,
				OriginalRaw:   []byte(`{"data": "some data"}`),
			})
			require.NoError(t, err)
		})

		t.Run("upsert failed for status code != 200", func(t *testing.T) {
			filter := crud.Filter{
				MongoQuery: map[string]any{
					"_pk.key1": "12345",
					"_pk.key2": "98765",
				},
			}

			gock.NewGockScope(t, "http://example.com/crud/", http.MethodPost, "upsert-one").
				AddMatcher(gock.CrudQueryMatcher(t, gock.Filter(filter))).
				Reply(500).JSON(map[string]any{"error": "Internal Server Error"})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL: "http://example.com/crud/",
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Write,
				OriginalRaw:   []byte(`{"data": "some data"}`),
			})
			require.Error(t, err)
		})

		t.Run("successful insert with insertOnly set to true", func(t *testing.T) {
			gock.NewGockScope(t, "http://example.com/crud/", http.MethodPost, "").
				Reply(200).JSON(map[string]any{})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL:        "http://example.com/crud/",
					InsertOnly: true,
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Write,
				OriginalRaw:   []byte(`{"data": "some data"}`),
			})
			require.NoError(t, err)
		})

		t.Run("failure insert with insertOnly set to true for status code != 200", func(t *testing.T) {
			gock.NewGockScope(t, "http://example.com/crud/", http.MethodPost, "").
				Reply(500).JSON(map[string]any{"error": "Internal Server Error"})

			w, err := NewWriter[entities.PipelineEvent](
				&Config{
					URL:        "http://example.com/crud/",
					InsertOnly: true,
				},
			)
			require.NoError(t, err)

			err = w.WriteData(t.Context(), &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "key1", Value: "12345"},
					{Key: "key2", Value: "98765"},
				},
				OperationType: entities.Write,
				OriginalRaw:   []byte(`{"data": "some data"}`),
			})
			require.Error(t, err)
		})
	})
}
