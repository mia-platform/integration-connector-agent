// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package sinks

import (
	"context"
	"errors"

	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrEmptyID = errors.New("id is empty")
)

type DataWithIdentifier interface {
	GetPrimaryKeys() entities.PkFields
	Operation() entities.Operation
}

// Sink interface abstract the implementation of an integration pipeline target. The concrete implementation has
// to know how to write and delete a Data.
type Sink[Data DataWithIdentifier] interface {
	// WriteData will save the Data to the destination configured in the Writer.
	// Data will have the operation to perform (write, delete) and the data to save,
	// which can be used based on the sink type.
	WriteData(ctx context.Context, data Data) error

	Close(ctx context.Context) error
}

const (
	Mongo          = "mongo"
	CRUDService    = "crud-service"
	ConsoleCatalog = "console-catalog"
	Kafka          = "kafka"

	// Fake is a fake writer used for testing purposes
	Fake = "fake"
)
