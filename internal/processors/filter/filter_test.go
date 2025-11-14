// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package filter

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		config Config
		event  entities.PipelineEvent

		expectedResult      entities.PipelineEvent
		expectedNewError    string
		expectedResultError string
	}{
		"fails to create new filter processor": {
			config: Config{
				CELExpression: `foo == eventType`,
			},

			expectedNewError: "ERROR: <input>:1:1: undeclared reference to 'foo' (in container '')\n | foo == eventType\n | ^",
		},
		"check on event type": {
			config: Config{
				CELExpression: `eventType == "event-type"`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type"}`),
			},

			expectedResult: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type"}`),
			},
		},
		"check if event type starts with": {
			config: Config{
				CELExpression: `eventType.startsWith("event-")`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type"}`),
			},

			expectedResult: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type"}`),
			},
		},
		"fails check if event type not starts with": {
			config: Config{
				CELExpression: `eventType.startsWith("not-correct")`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type"}`),
			},

			expectedResultError: entities.ErrDiscardEvent.Error(),
		},
		"check on multiple event type": {
			config: Config{
				CELExpression: `eventType in ["event-type", "event-type-2", "event-type-3"]`,
			},
			event: &entities.Event{
				Type:        "event-type-2",
				OriginalRaw: []byte(`{"type":"event-type-2"}`),
			},

			expectedResult: &entities.Event{
				Type:        "event-type-2",
				OriginalRaw: []byte(`{"type":"event-type-2"}`),
			},
		},
		"check event not set": {
			config: Config{
				CELExpression: `eventType in ["event-type", "event-type-2", "event-type-3"]`,
			},
			event: &entities.Event{
				Type:        "event-to-filter",
				OriginalRaw: []byte(`{"type":"event-to-filter"}`),
			},

			expectedResultError: entities.ErrDiscardEvent.Error(),
		},
		"check on event content": {
			config: Config{
				CELExpression: `data.fields.id == "my-id"`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type","fields":{"id": "my-id"}}`),
			},

			expectedResult: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type","fields":{"id": "my-id"}}`),
			},
		},
		"event to filter on id": {
			config: Config{
				CELExpression: `data.fields.id != "my-id"`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type","fields":{"id": "my-id"}}`),
			},

			expectedResultError: entities.ErrDiscardEvent.Error(),
		},
		"not an expression": {
			config: Config{
				CELExpression: `"Hello world! The event type is " + eventType`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type","fields":{"id": "my-id"}}`),
			},

			expectedResult: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`{"type":"event-type","fields":{"id": "my-id"}}`),
			},
		},
		"fails if event is not a JSON": {
			config: Config{
				CELExpression: `true`,
			},
			event: &entities.Event{
				Type:        "event-type",
				OriginalRaw: []byte(`is not a json`),
			},

			expectedResultError: "invalid character 'i' looking for beginning of value",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			filter, err := New(tc.config)
			if tc.expectedNewError != "" {
				require.EqualError(t, err, tc.expectedNewError)
				return
			}
			require.NoError(t, err)

			data, err := filter.Process(tc.event)
			if tc.expectedResultError != "" {
				require.EqualError(t, err, tc.expectedResultError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedResult, data)
		})
	}
}
