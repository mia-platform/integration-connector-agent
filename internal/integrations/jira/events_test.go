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

package jira

import (
	"fmt"
	"testing"

	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestEvent(t *testing.T) {
	testCases := map[string]struct {
		rawData string

		expectError           string
		expectedID            string
		expectedOperationType entities.Operation
	}{
		"issue_created": {
			rawData: fmt.Sprintf(`{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "%s"}`, issueCreated),

			expectedID:            "my-id",
			expectedOperationType: entities.Write,
		},
		"issue_updated": {
			rawData: fmt.Sprintf(`{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "%s"}`, issueUpdated),

			expectedID:            "my-id",
			expectedOperationType: entities.Write,
		},
		"issue_deleted": {
			rawData: fmt.Sprintf(`{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "%s"}`, issueDeleted),

			expectedID:            "my-id",
			expectedOperationType: entities.Delete,
		},
		"unsupported_event": {
			rawData: `{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "unsupported"}`,

			expectError: fmt.Sprintf("%s: %s", ErrUnsupportedWebhookEvent, "unsupported"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := getPipelineEvent([]byte(tc.rawData))
			if tc.expectError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)

				parsed := gjson.ParseBytes([]byte(tc.rawData))
				require.Equal(t, &entities.Event{
					ID:            tc.expectedID,
					OperationType: tc.expectedOperationType,

					OriginalRaw:    []byte(tc.rawData),
					OriginalParsed: parsed,
				}, event)
			}
		})
	}
}
