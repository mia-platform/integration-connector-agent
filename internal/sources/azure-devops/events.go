// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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
