// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

	InventoryEventStorageType  = "storage.googleapis.com/Bucket"
	InventoryEventFunctionType = "run.googleapis.com/Service"
)

type InventoryEventBuilder[T IInventoryEvent] struct{}

type IInventoryEvent interface {
	Name() string
	AssetType() string
	Ancestors() []string
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
