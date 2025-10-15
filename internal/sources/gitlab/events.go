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

package gitlab

import (
	"github.com/gofiber/fiber/v2"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	gitlabEventHeader = "X-Gitlab-Event"

	// Project events
	projectEvent = "Project Hook"

	// Merge request events
	mergeRequestEvent = "Merge Request Hook"

	// Pipeline events
	pipelineEvent = "Pipeline Hook"

	// Release events
	releaseEvent = "Release Hook"

	// Push events
	pushEvent = "Push Hook"

	// Tag events
	tagEvent = "Tag Push Hook"

	// Issue events
	issueEvent = "Issue Hook"

	// Note (comment) events
	noteEvent = "Note Hook"

	// Wiki page events
	wikiPageEvent = "Wiki Page Hook"

	// Deployment events
	deploymentEvent = "Deployment Hook"

	// Job events
	jobEvent = "Job Hook"

	// Build events (legacy)
	buildEvent = "Build Hook"

	// System hook events
	systemEvent = "System Hook"

	// Feature flag events
	featureFlagEvent = "Feature Flag Hook"

	// Repository update events
	repositoryUpdateEvent = "Repository Update Hook"

	// Member events
	memberEvent = "Member Hook"

	// Subgroup events
	subgroupEvent = "Subgroup Hook"

	// Security events
	vulnerabilityEvent = "Vulnerability Hook"

	// Archive/Unarchive events
	archiveEvent = "Archive Hook"
)

var SupportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		// Project events
		projectEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project.id"),
		},

		// Merge request events
		mergeRequestEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Pipeline events
		pipelineEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Release events
		releaseEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("id"),
		},

		// Push events
		pushEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project.id"),
		},

		// Tag events
		tagEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project.id"),
		},

		// Issue events
		issueEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Note (comment) events
		noteEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Wiki page events
		wikiPageEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Deployment events
		deploymentEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("deployment_id"),
		},

		// Job events
		jobEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("build_id"),
		},

		// Build events (legacy)
		buildEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("build_id"),
		},

		// System hook events
		systemEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project_id"),
		},

		// Feature flag events
		featureFlagEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Repository update events
		repositoryUpdateEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project_id"),
		},

		// Member events
		memberEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("user_id"),
		},

		// Subgroup events
		subgroupEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("group_id"),
		},

		// Security events
		vulnerabilityEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("object_attributes.id"),
		},

		// Archive/Unarchive events
		archiveEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project_id"),
		},
	},
	GetEventType: func(data webhook.EventTypeParam) string {
		return data.Headers.Get(gitlabEventHeader)
	},
	PayloadKey: webhook.ContentTypeConfig{
		fiber.MIMEApplicationForm: "payload",
	},
}
