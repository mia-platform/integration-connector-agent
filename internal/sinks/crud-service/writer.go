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

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
)

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
)

type Writer[T entities.PipelineEvent] struct {
	url        string
	insertOnly bool
	client     crudclient[T]
}

func NewWriter[T entities.PipelineEvent](config *Config) (sinks.Sink[T], error) {
	client, err := newCRUDClient[T](config.URL, config.PrimaryKey)
	if err != nil {
		return nil, err
	}

	return &Writer[T]{
		url:        config.URL,
		insertOnly: config.InsertOnly,
		client:     client,
	}, nil
}

func (w *Writer[T]) WriteData(ctx context.Context, data T) error {
	if w.insertOnly {
		return w.client.Insert(ctx, data)
	}

	switch data.Operation() {
	case entities.Delete:
		return w.client.Delete(ctx, data)
	case entities.Write:
		return w.client.Upsert(ctx, data)
	default:
		return ErrUnsupportedOperation
	}
}
