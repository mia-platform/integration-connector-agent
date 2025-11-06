// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gitlab

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/tidwall/gjson"
)

const (
	gitlabEventHeader = "X-Gitlab-Event"

	// Deployment events
	deploymentEvent = "Deployment Hook"

	// Feature flag events
	featureFlagEvent = "Feature Flag Hook"

	// Merge request events
	mergeRequestEvent = "Merge Request Hook"

	// Pipeline events
	pipelineEvent = "Pipeline Hook"

	// Release events
	releaseEvent = "Release Hook"

	// Tag events
	tagEvent = "Tag Push Hook"

	// Security events
	vulnerabilityEvent = "Vulnerability Hook"

	// Issue events
	issueEvent = "Issue Hook"

	// Project events
	projectEvent = "Project Hook"

	// Subgroup events
	subgroupEvent = "Subgroup Hook"
)

func supportedEvents(baseURL string) *webhook.Events {
	return &webhook.Events{
		Supported: map[string]webhook.Event{
			// Project events
			projectEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("project_id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "id", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Merge request events
			mergeRequestEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("object_attributes.id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "id", Value: value},
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Pipeline events
			pipelineEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("object_attributes.id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "id", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Release events
			releaseEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("tag").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "tagName", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Tag events
			tagEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("ref").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "ref", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Issue events
			issueEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("object_attributes.id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "id", Value: value},
						{Key: "objectKind", Value: parsedData.Get("object_kind").String()},
						{Key: "eventType", Value: parsedData.Get("event_type").String()},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Deployment events
			deploymentEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("deployment_id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "id", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Feature flag events
			featureFlagEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("object_attributes.id").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("project.id").String()},
						{Key: "id", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Subgroup events
			subgroupEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("full_path").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "fullPath", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},

			// Security events
			vulnerabilityEvent: {
				Operation: entities.Write,
				GetFieldID: func(parsedData gjson.Result) entities.PkFields {
					value := parsedData.Get("object_attributes.url").String()
					if value == "" {
						return nil
					}

					return entities.PkFields{
						{Key: "projectId", Value: parsedData.Get("object_attributes.project_id").String()},
						{Key: "vulnerabilityUrl", Value: value},
						{Key: "url", Value: baseURL},
					}
				},
			},
		},

		GetEventType: func(data webhook.EventTypeParam) string {
			data.Headers.Get(gitlabEventHeader)
			return data.Headers.Get(gitlabEventHeader)
		},
	}
}
