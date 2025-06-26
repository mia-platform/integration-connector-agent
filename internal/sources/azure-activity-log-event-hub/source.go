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

package azureactivitylogeventhub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	azureeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-event-hub"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/sirupsen/logrus"
)

type ActivityLogEventData struct {
	Records []*ActivityLogEventRecord `json:"records"`
}

type ActivityLogEventRecord struct {
	RoleLocation    string                    `json:"RoleLocation"`   //nolint:tagliatelle
	Stamp           string                    `json:"Stamp"`          //nolint:tagliatelle
	ReleaseVersion  string                    `json:"ReleaseVersion"` //nolint:tagliatelle
	Time            string                    `json:"time"`
	ResourceID      string                    `json:"resourceId"`
	OperationName   string                    `json:"operationName"`
	Category        string                    `json:"category"`
	ResultType      string                    `json:"resultType"`
	ResultSignature string                    `json:"resultSignature"`
	DurationMs      string                    `json:"durationMs"`
	CallerIPAddress string                    `json:"callerIpAddress"`
	CorrelationID   string                    `json:"correlationId"`
	Identity        *ActivityLogEventIdentity `json:"identity"`
	Level           string                    `json:"level"`
	Properties      map[string]any            `json:"properties"`
}

type ActivityLogEventIdentity struct {
	Authorization *ActivityLogAuthorization `json:"authorization"`
	Claims        map[string]string         `json:"claims"`
}

type ActivityLogAuthorization struct {
	Scope    string                `json:"scope"`
	Action   string                `json:"action"`
	Evidence AuthorizationEvidence `json:"evidence"`
}

type AuthorizationEvidence struct {
	Role                string `json:"role"`
	RoleAssignmentScope string `json:"roleAssignmentScope"`
	RoleAssignmentID    string `json:"roleAssignmentId"`
	RoleDefinitionID    string `json:"roleDefinitionId"`
	PrincipalID         string `json:"principalId"`
	PrincipalType       string `json:"principalType"`
}

func AddSource(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger) error {
	eventHubConfig, err := configFromGeneric(cfg, pg)
	if err != nil {
		return err
	}

	return azureeventhub.SetupEventHub(ctx, eventHubConfig, pg, logger)
}

func configFromGeneric(cfg config.GenericConfig, pg pipeline.IPipelineGroup) (*azureeventhub.Config, error) {
	eventHubConfig, err := config.GetConfig[*azureeventhub.Config](cfg)
	if err != nil {
		return nil, err
	}

	eventHubConfig.EventConsumer = activityLogConsumer(pg)
	return eventHubConfig, nil
}

func activityLogConsumer(pg pipeline.IPipelineGroup) azureeventhub.EventConsumer {
	return func(eventData *azeventhubs.ReceivedEventData) error {
		activityLogEventData := new(ActivityLogEventData)
		if err := json.Unmarshal(eventData.Body, activityLogEventData); err != nil {
			return fmt.Errorf("failed to read activity log event data: %w", err)
		}

		for _, record := range activityLogEventData.Records {
			if event := pipelineEventFromRecord(record); event != nil {
				pg.AddMessage(event)
			}
		}

		return nil
	}
}

func pipelineEventFromRecord(record *ActivityLogEventRecord) *entities.Event {
	rawRecord, err := json.Marshal(record)
	if err != nil {
		return nil
	}

	return &entities.Event{
		PrimaryKeys: entities.PkFields{
			{
				Key:   "resourceId",
				Value: strings.ToLower(record.ResourceID),
			},
		},
		OperationType: eventOperationTypeFromRecord(record),
		Type:          eventTypeFromRecord(record),
		OriginalRaw:   rawRecord,
	}
}

func eventOperationTypeFromRecord(record *ActivityLogEventRecord) entities.Operation {
	if strings.HasSuffix(strings.ToLower(record.OperationName), "delete") ||
		strings.HasSuffix(strings.ToLower(record.OperationName), "delete/action") {
		return entities.Delete
	}

	return entities.Write
}

func eventTypeFromRecord(record *ActivityLogEventRecord) string {
	eventType := fmt.Sprintf("%s:%s", strings.ToLower(record.OperationName), strings.ToLower(record.Category))
	eventType = strings.ReplaceAll(eventType, "microsoft", "azure")
	eventType = strings.ReplaceAll(eventType, ".", ":")
	eventType = strings.ReplaceAll(eventType, "/", ":")
	return eventType
}
