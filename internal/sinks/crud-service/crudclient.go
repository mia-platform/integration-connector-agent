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
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/mia-platform/go-crud-service-client"
)

type iClient[T entities.PipelineEvent] interface {
	Upsert(ctx context.Context, event T) error
	Delete(ctx context.Context, event T) error
}

type client[T entities.PipelineEvent] struct {
	c crud.CrudClient[any]
}

func newCRUDClient[T entities.PipelineEvent](url string) (*client[T], error) {
	c, err := crud.NewClient[any](crud.ClientOptions{
		BaseURL: url,
	})
	if err != nil {
		return nil, err
	}

	return &client[T]{c: c}, nil
}

func (c *client[T]) Upsert(ctx context.Context, event T) error {
	m := map[string]any{}
	if err := json.Unmarshal(event.Data(), &m); err != nil {
		return err
	}

	pks := event.GetPrimaryKeys().Map()
	for k, v := range pks {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}

	upsert := crud.UpsertBody{Set: m}
	_, err := c.c.UpsertOne(ctx, upsert, crud.Options{
		Filter: crud.Filter{
			Fields: pks,
		},
	})
	return err
}
func (c *client[T]) Delete(ctx context.Context, event T) error {
	_, err := c.c.DeleteMany(ctx, crud.Options{
		Filter: crud.Filter{
			Fields: event.GetPrimaryKeys().Map(),
		},
	})
	return err
}
