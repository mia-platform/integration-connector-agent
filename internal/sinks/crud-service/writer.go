// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

func (w *Writer[T]) Close(_ context.Context) error {
	return nil
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
