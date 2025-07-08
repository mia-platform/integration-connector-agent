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

package commons

import (
	"testing"

	azureactivitylogeventhubevents "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub/events"
	"github.com/stretchr/testify/assert"
)

func TestEventIsForResourceType(t *testing.T) {
	testCases := map[string]struct {
		event        *azureactivitylogeventhubevents.ActivityLogEventRecord
		resourceType string
		expected     bool
	}{
		"event update tags for functions": {
			event: &azureactivitylogeventhubevents.ActivityLogEventRecord{
				OperationName: "MICROSOFT.RESOURCES/TAGS/WRITE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: "microsoft.web/sites",
			expected:     true,
		},
		"event for functions": {
			event: &azureactivitylogeventhubevents.ActivityLogEventRecord{
				OperationName: "MICROSOFT.WEB/SITES/WRITE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: "microsoft.web/sites",
			expected:     true,
		},
		"delete event for functions": {
			event: &azureactivitylogeventhubevents.ActivityLogEventRecord{
				OperationName: "MICROSOFT.WEB/SITES/DELETE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: "microsoft.web/sites",
			expected:     false,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			result := EventIsForSource(test.event, test.resourceType)
			assert.Equal(t, test.expected, result)
		})
	}
}
