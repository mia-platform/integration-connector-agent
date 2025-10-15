// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package awssqsevents

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

	CloudTrailEventStorageType  = "s3.amazonaws.com"
	CloudTrailEventFunctionType = "lambda.amazonaws.com"
)

type IEvent interface {
	ResourceName() (string, error)
	EventSource() string
	GetRegion() string
	AccountID() string
	Operation() (entities.Operation, error)
	EventType() string
}

func NewCloudTrailEventBuilder[T IEvent]() entities.EventBuilder {
	return &CloudTrailEventBuilder[T]{}
}

type CloudTrailEventBuilder[T IEvent] struct{}

func (b CloudTrailEventBuilder[T]) GetPipelineEvent(_ context.Context, data []byte) (entities.PipelineEvent, error) {
	var rawEvent T
	if err := json.Unmarshal(data, &rawEvent); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedEvent, err.Error())
	}

	pk, err := b.primaryKeys(rawEvent)
	if err != nil {
		return nil, fmt.Errorf("error getting primary keys: %w", err)
	}

	op, err := b.operationType(rawEvent)
	if err != nil {
		return nil, fmt.Errorf("error getting operation type: %w", err)
	}

	return &entities.Event{
		PrimaryKeys:   pk,
		OperationType: op,
		Type:          b.eventType(rawEvent),
		OriginalRaw:   data,
	}, nil
}

func (CloudTrailEventBuilder[T]) primaryKeys(event T) (entities.PkFields, error) {
	resourceName, err := event.ResourceName()
	if err != nil {
		return nil, err
	}
	return entities.PkFields{
		{Key: "resourceName", Value: resourceName},
		{Key: "eventSource", Value: event.EventSource()},
	}, nil
}

func (CloudTrailEventBuilder[T]) operationType(event T) (entities.Operation, error) {
	return event.Operation()
}

func (CloudTrailEventBuilder[T]) eventType(event T) string {
	return event.EventType()
}
