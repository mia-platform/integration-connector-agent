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

package sinks

import (
	"context"
	"errors"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
)

var (
	ErrEmptyID = errors.New("id is empty")
)

type DataWithIdentifier interface {
	GetID() string
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
	Mongo = "mongo"
	Kafka = "kafka"

	// Fake is a fake writer used for testing purposes
	Fake = "fake"
)
