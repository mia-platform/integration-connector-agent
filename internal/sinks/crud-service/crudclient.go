// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

type crudclient[T entities.PipelineEvent] interface {
	Upsert(ctx context.Context, event T) error
	Delete(ctx context.Context, event T) error
	Insert(ctx context.Context, event T) error
}

type client[T entities.PipelineEvent] struct {
	c             crud.CrudClient[any]
	pkFieldPrefix string
}

func newCRUDClient[T entities.PipelineEvent](url string, pkFieldPrefix string) (crudclient[T], error) {
	c, err := crud.NewClient[any](crud.ClientOptions{
		BaseURL: url,
	})
	if err != nil {
		return nil, err
	}

	return &client[T]{
		c:             c,
		pkFieldPrefix: pkFieldPrefix,
	}, nil
}

func (c *client[T]) Upsert(ctx context.Context, event T) error {
	m, err := c.prepareData(event)
	if err != nil {
		return err
	}

	_, err = c.c.UpsertOne(
		ctx,
		crud.UpsertBody{Set: m},
		crud.Options{
			Filter: crud.Filter{
				MongoQuery: c.prepareMongoQueryFilter(event),
			},
		},
	)
	return err
}

func (c *client[T]) Delete(ctx context.Context, event T) error {
	_, err := c.c.DeleteMany(ctx, crud.Options{
		Filter: crud.Filter{
			MongoQuery: c.prepareMongoQueryFilter(event),
		},
	})
	return err
}

func (c *client[T]) Insert(ctx context.Context, event T) error {
	data, err := c.prepareData(event)
	if err != nil {
		return err
	}

	_, err = c.c.Create(ctx, data, crud.Options{})
	return err
}

func (c *client[T]) prepareData(event T) (map[string]any, error) {
	data := map[string]any{}
	if err := json.Unmarshal(event.Data(), &data); err != nil {
		return nil, err
	}

	data[c.pkFieldPrefix] = event.GetPrimaryKeys().Map()
	return data, nil
}

func (c *client[T]) prepareMongoQueryFilter(event T) map[string]any {
	pks := event.GetPrimaryKeys().Map()

	filter := make(map[string]any, len(pks))
	for k, v := range pks {
		filter[c.pkKey(k)] = v
	}
	return filter
}

func (c *client[T]) pkKey(key string) string {
	if c.pkFieldPrefix == "" {
		return key
	}
	return c.pkFieldPrefix + "." + key
}
