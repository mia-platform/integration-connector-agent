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

package azuredevops

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/tidwall/gjson"
)

const (
	repositoryCreated = "git.repo.created"
	repositoryRenamed = "git.repo.renamed"
	repositoryDeleted = "git.repo.deleted"

	primaryKeyFieldPath        = "resource.repository.id"
	deletedPrimaryKeyFieldPath = "resource.repositoryId"
	eventTypeFieldPath         = "eventType"
)

var supportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		repositoryCreated: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyByPath(primaryKeyFieldPath),
		},
		repositoryRenamed: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyByPath(primaryKeyFieldPath),
		},
		repositoryDeleted: {
			Operation:  entities.Delete,
			GetFieldID: getPrimaryKeyByPath(deletedPrimaryKeyFieldPath),
		},
	},
	GetEventType: webhook.GetEventTypeByPath(eventTypeFieldPath),
}

func getPrimaryKeyByPath(path string) func(parsedData gjson.Result) entities.PkFields {
	return func(parsedData gjson.Result) entities.PkFields {
		value := parsedData.Get(path).String()
		if value == "" {
			return nil
		}

		return entities.PkFields{
			{
				Key:   "repositoryId",
				Value: value,
			},
			{
				Key:   "type",
				Value: "repository",
			},
		}
	}
}
