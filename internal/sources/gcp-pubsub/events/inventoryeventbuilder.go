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

package gcppubsubevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrMalformedEvent = errors.New("malformed event")
)

const (
	RealtimeSyncEventType = "sync-event"
	ImportEventType       = "import-event"

	InventoryEventStorageType = "storage.googleapis.com/Bucket"
)

type InventoryEventBuilder[T IInventoryEvent] struct {
	builtEventType string
}

type IInventoryEvent interface {
	Name() string
	AssetType() string
	Operation() entities.Operation
	EventType() string
}

func NewInventoryEventBuilder[T IInventoryEvent]() entities.EventBuilder {
	return &InventoryEventBuilder[T]{}
}

func (b *InventoryEventBuilder[T]) GetPipelineEvent(_ context.Context, data []byte) (entities.PipelineEvent, error) {
	var event T
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedEvent, err.Error())
	}

	return &entities.Event{
		PrimaryKeys: entities.PkFields{
			{Key: "resourceName", Value: event.Name()},
			{Key: "resourceType", Value: event.AssetType()},
		},
		OperationType: event.Operation(),
		Type:          event.EventType(),
		OriginalRaw:   data,
	}, nil
}
