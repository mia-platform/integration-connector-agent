// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package console

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"

	"github.com/tidwall/gjson"
)

const (
	projectCreatedEvent     = "project_created"
	serviceCreatedEvent     = "service_created"
	tagCreatedEvent         = "tag_created"
	configurationSavedEvent = "configuration_saved"

	tenantIDEventPath     = "payload.tenantId"
	projectIDEventPath    = "payload.projectId"
	serviceNameEventPath  = "payload.serviceName"
	tagNameEventPath      = "payload.tagName"
	revisionNameEventPath = "payload.revisionName"

	tenantIDKey     = "tenantId"
	projectIDKey    = "projectId"
	serviceNameKey  = "serviceName"
	revisionNameKey = "revisionName"
)

var SupportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		projectCreatedEvent: {
			Operation: entities.Write,
			GetFieldID: func(parsedData gjson.Result) entities.PkFields {
				projectID := parsedData.Get(projectIDEventPath).String()
				tenantID := parsedData.Get(tenantIDEventPath).String()

				return entities.PkFields{
					entities.PkField{Key: tenantIDKey, Value: tenantID},
					entities.PkField{Key: projectIDKey, Value: projectID},
				}
			},
		},
		serviceCreatedEvent: {
			Operation: entities.Write,
			GetFieldID: func(parsedData gjson.Result) entities.PkFields {
				projectID := parsedData.Get(projectIDEventPath).String()
				serviceName := parsedData.Get(serviceNameEventPath).String()
				tenantID := parsedData.Get(tenantIDEventPath).String()

				return entities.PkFields{
					entities.PkField{Key: tenantIDKey, Value: tenantID},
					entities.PkField{Key: projectIDKey, Value: projectID},
					entities.PkField{Key: serviceNameKey, Value: serviceName},
				}
			},
		},
		configurationSavedEvent: {
			Operation: entities.Write,
			GetFieldID: func(parsedData gjson.Result) entities.PkFields {
				tenantID := parsedData.Get(tenantIDEventPath).String()
				projectID := parsedData.Get(projectIDEventPath).String()
				revisionName := parsedData.Get(revisionNameEventPath).String()

				return entities.PkFields{
					entities.PkField{Key: tenantIDKey, Value: tenantID},
					entities.PkField{Key: projectIDKey, Value: projectID},
					entities.PkField{Key: revisionNameKey, Value: revisionName},
				}
			},
		},
		tagCreatedEvent: {
			Operation: entities.Write,
			GetFieldID: func(parsedData gjson.Result) entities.PkFields {
				tenantID := parsedData.Get(tenantIDEventPath).String()
				projectID := parsedData.Get(projectIDEventPath).String()
				tagName := parsedData.Get(tagNameEventPath).String()

				return entities.PkFields{
					entities.PkField{Key: tenantIDKey, Value: tenantID},
					entities.PkField{Key: projectIDKey, Value: projectID},
					entities.PkField{Key: "tagName", Value: tagName},
				}
			},
		},
	},
	GetEventType: webhook.GetEventTypeByPath(webhookEventPath),
}
