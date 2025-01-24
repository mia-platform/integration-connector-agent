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

package console

import (
	"github.com/mia-platform/integration-connector-agent/internal/entities"
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

var DefaultSupportedEvents = webhook.Events{
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
	EventTypeFieldPath: webhookEventPath,
}
