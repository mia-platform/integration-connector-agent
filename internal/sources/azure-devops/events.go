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

package azuredevops

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	repositoryCreated = "azuredevops:repository_created"
	repositoryRenamed = "azuredevops:repository_renamed"
	repositoryDeleted = "azuredevops:repository_deleted"

	primaryKeyFieldPath = ""
	eventTypeFieldPath  = ""
)

var supportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		repositoryCreated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(primaryKeyFieldPath),
		},
		repositoryRenamed: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(primaryKeyFieldPath),
		},
		repositoryDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(primaryKeyFieldPath),
		},
	},
	GetEventType: webhook.GetEventTypeByPath(eventTypeFieldPath),
}
