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
	"strings"

	azureactivitylogeventhubevents "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub/events"
)

func EventIsForSource(event *azureactivitylogeventhubevents.ActivityLogEventRecord, resourceType string) bool {
	eventSource := strings.ToLower(event.OperationName)
	resourceID := strings.ToLower(event.ResourceID)

	return eventSource == resourceType+"/write" ||
		(eventSource == "microsoft.resources/tags/write" && strings.Contains(resourceID, resourceType))
}
