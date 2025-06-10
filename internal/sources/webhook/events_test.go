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

package webhook

import (
	"fmt"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	logger, _ := test.NewNullLogger()

	testCases := map[string]struct {
		rawData string
		events  *Events

		expectError           string
		expectedPk            entities.PkFields
		expectedType          string
		expectedOperationType entities.Operation
	}{
		"without id in the event": {
			rawData: `{"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Write,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectError: "missing id field in event: my-event",
		},
		"supported write event": {
			rawData: `{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Write,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectedPk:            entities.PkFields{{Key: "issue.id", Value: "my-id"}},
			expectedType:          "my-event",
			expectedOperationType: entities.Write,
		},
		"supported delete event": {
			rawData: `{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						GetFieldID: GetPrimaryKeyByPath("issue.id"),
						Operation:  entities.Delete,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectedOperationType: entities.Delete,
			expectedType:          "my-event",
			expectedPk:            entities.PkFields{{Key: "issue.id", Value: "my-id"}},
		},
		"unsupported_event": {
			rawData: `{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "unsupported"}`,
			events: &Events{
				EventTypeFieldPath: "webhookEvent",
			},

			expectError:  fmt.Sprintf("%s: %s", ErrUnsupportedWebhookEvent, "unsupported"),
			expectedType: "unsupported",
		},
		"with custom GetFieldID": {
			rawData: `{"issue":{"tag":"my-id","projectId":"prj-1","parentId":"my-parent-id"},"webhookEvent": "my-event"}`,
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
				EventTypeFieldPath: "webhookEvent",
			},

			expectedPk:            entities.PkFields{{Key: "parent", Value: "my-parent-id"}, {Key: "project", Value: "prj-1"}},
			expectedType:          "my-event",
			expectedOperationType: entities.Write,
		},
		"without GetFieldID": {
			rawData: `{"issue":{"tag":"my-id","projectId":"prj-1","parentId":"my-parent-id"},"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {},
				},
				EventTypeFieldPath: "webhookEvent",
			},

			expectError: fmt.Sprintf("%s: my-event missing GetFieldID function", ErrUnsupportedWebhookEvent),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := tc.events.getPipelineEvent(logrus.NewEntry(logger), []byte(tc.rawData))
			if tc.expectError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)

				require.Equal(t, &entities.Event{
					PrimaryKeys:   tc.expectedPk,
					OperationType: tc.expectedOperationType,
					Type:          tc.expectedType,

					OriginalRaw: []byte(tc.rawData),
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
