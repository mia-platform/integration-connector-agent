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

package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestEvent(t *testing.T) {
	logger, _ := test.NewNullLogger()

	testCases := map[string]struct {
		requestInfo RequestInfo
		events      *Events

		expectError           string
		expectedPk            entities.PkFields
		expectedType          string
		expectedOperationType entities.Operation
	}{
		"without id in the event": {
			requestInfo: RequestInfo{
				data: []byte(`{"webhookEvent": "my-event"}`),
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Write,
					},
				},
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},
			expectError: "missing id field in event: my-event",
		},
		"supported write event": {
			requestInfo: RequestInfo{
				data: []byte(`{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`),
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Write,
					},
				},
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},
			expectedPk:            entities.PkFields{{Key: "issue.id", Value: "my-id"}},
			expectedType:          "my-event",
			expectedOperationType: entities.Write,
		},
		"supported delete event": {
			requestInfo: RequestInfo{
				data: []byte(`{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`),
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Delete,
					},
				},
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},
			expectedOperationType: entities.Delete,
			expectedType:          "my-event",
			expectedPk:            entities.PkFields{{Key: "issue.id", Value: "my-id"}},
		},
		"unsupported_event": {
			requestInfo: RequestInfo{
				data: []byte(`{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "unsupported"}`),
			},
			events: &Events{
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},

			expectError:  fmt.Sprintf("%s: %s", ErrUnsupportedWebhookEvent, "unsupported"),
			expectedType: "unsupported",
		},
		"with custom GetFieldID": {
			requestInfo: RequestInfo{
				data: []byte(`{"issue":{"tag":"my-id","projectId":"prj-1","parentId":"my-parent-id"},"webhookEvent": "my-event"}`),
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: func(parsedData gjson.Result) entities.PkFields {
							return entities.PkFields{
								{Key: "parent", Value: parsedData.Get("issue.parentId").String()},
								{Key: "project", Value: parsedData.Get("issue.projectId").String()},
							}
						},
					},
				},
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},

			expectedPk:            entities.PkFields{{Key: "parent", Value: "my-parent-id"}, {Key: "project", Value: "prj-1"}},
			expectedType:          "my-event",
			expectedOperationType: entities.Write,
		},
		"without GetFieldID": {
			requestInfo: RequestInfo{
				data: []byte(`{"issue":{"tag":"my-id","projectId":"prj-1","parentId":"my-parent-id"},"webhookEvent": "my-event"}`),
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {},
				},
				GetEventType: GetEventTypeByPath("webhookEvent"),
			},

			expectError: fmt.Sprintf("%s: my-event missing GetFieldID function", ErrUnsupportedWebhookEvent),
		},
		"with event id from header": {
			requestInfo: RequestInfo{
				data:    []byte(`{"issue":{"id":"my-id"}}`),
				headers: http.Header{"Event-Type": []string{"my-event"}},
			},
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Write,
					},
				},
				GetEventType: func(data EventTypeParam) string {
					return data.Headers.Get("Event-Type")
				},
			},

			expectedPk:            entities.PkFields{{Key: "issue.id", Value: "my-id"}},
			expectedType:          "my-event",
			expectedOperationType: entities.Write,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := tc.events.getPipelineEvent(logrus.NewEntry(logger), tc.requestInfo)
			if tc.expectError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)

				require.Equal(t, &entities.Event{
					PrimaryKeys:   tc.expectedPk,
					OperationType: tc.expectedOperationType,
					Type:          tc.expectedType,

					OriginalRaw: getExpectedWebhookPayload(tc.requestInfo, tc.events),
				}, event)
			}
		})
	}
}

func TestGetPrimaryKeyByPath(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		getFieldID := GetPrimaryKeyByPath("issue.id")
		parsed := gjson.ParseBytes([]byte(`{"issue": {"id": "my-id"}}`))

		require.Equal(t, entities.PkFields{{Key: "issue.id", Value: "my-id"}}, getFieldID(parsed))
	})

	t.Run("empty", func(t *testing.T) {
		getFieldID := GetPrimaryKeyByPath("issue.id")
		parsed := gjson.ParseBytes([]byte(`{}`))

		require.Nil(t, getFieldID(parsed))
	})
}

func TestGetEventTypeByPath(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		getEventType := GetEventTypeByPath("event.type")
		parsed := gjson.ParseBytes([]byte(`{"event": {"type": "my-type"}}`))
		require.Equal(t, "my-type", getEventType(EventTypeParam{
			Data: parsed,
		}))
	})
}

// getExpectedWebhookPayload determines if eventType should be injected based on the webhook configuration
func getExpectedWebhookPayload(requestInfo RequestInfo, events *Events) []byte {
	// Get event type from the events configuration
	eventType := events.GetEventType(EventTypeParam{
		Data:    gjson.ParseBytes(requestInfo.data),
		Headers: requestInfo.headers,
	})

	// Check if the event type was extracted from the payload itself
	// by comparing with what GetEventType returns when called with empty headers
	eventTypeFromPayloadOnly := events.GetEventType(EventTypeParam{
		Data:    gjson.ParseBytes(requestInfo.data),
		Headers: http.Header{},
	})

	// Only inject eventType if it came from headers (not from payload)
	if eventType != "" && eventTypeFromPayloadOnly == "" {
		// Parse the existing JSON and add eventType field
		var jsonData map[string]interface{}
		if err := json.Unmarshal(requestInfo.data, &jsonData); err == nil {
			jsonData["eventType"] = eventType
			if enhancedBytes, err := json.Marshal(jsonData); err == nil {
				return enhancedBytes
			}
		}
	}

	// Return original data if no injection needed
	return requestInfo.data
}
