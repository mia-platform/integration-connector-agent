// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package jira

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	// issue events
	issueCreated = "jira:issue_created"
	issueUpdated = "jira:issue_updated"
	issueDeleted = "jira:issue_deleted"
	// issuelink events
	issueLinkCreated = "issuelink_created"
	issueLinkDeleted = "issuelink_deleted"
	// project events
	projectCreated         = "project_created"
	projectUpdated         = "project_updated"
	projectDeleted         = "project_deleted"
	projectSoftDeleted     = "project_soft_deleted"
	projectRestoredDeleted = "project_restored_deleted"
	// version events
	versionReleased   = "jira:version_released"
	versionUnreleased = "jira:version_unreleased"
	versionCreated    = "jira:version_created"
	versionUpdated    = "jira:version_updated"
	versionDeleted    = "jira:version_deleted"

	issueEventIDPath     = "issue.id"
	issueLinkEventIDPath = "issueLink.id"
	projectEventIDPath   = "project.id"
	versionEventIDPath   = "version.id"
)

var SupportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		issueCreated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(issueEventIDPath),
		},
		issueUpdated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(issueEventIDPath),
		},
		issueDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(issueEventIDPath),
		},
		issueLinkCreated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(issueLinkEventIDPath),
		},
		issueLinkDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(issueLinkEventIDPath),
		},
		projectCreated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(projectEventIDPath),
		},
		projectUpdated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(projectEventIDPath),
		},
		projectDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(projectEventIDPath),
		},
		projectSoftDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(projectEventIDPath),
		},
		projectRestoredDeleted: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(projectEventIDPath),
		},
		versionReleased: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(versionEventIDPath),
		},
		versionUnreleased: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(versionEventIDPath),
		},
		versionCreated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(versionEventIDPath),
		},
		versionUpdated: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath(versionEventIDPath),
		},
		versionDeleted: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath(versionEventIDPath),
		},
	},
	GetEventType: webhook.GetEventTypeByPath(webhookEventPath),
}
