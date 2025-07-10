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

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	azureeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-event-hub"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/sirupsen/logrus"
)

func AddSource(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger) error {
	eventHubConfig, err := configFromGeneric(cfg, pg)
	if err != nil {
		return err
	}

	return azureeventhub.SetupEventHub(ctx, eventHubConfig, pg, logger)
}

func configFromGeneric(cfg config.GenericConfig, pg pipeline.IPipelineGroup) (*azure.EventHubConfig, error) {
	eventHubConfig, err := config.GetConfig[*azure.EventHubConfig](cfg)
	if err != nil {
		return nil, err
	}

	eventHubConfig.EventConsumer = activityLogConsumer(pg)
	return eventHubConfig, nil
}

func activityLogConsumer(pg pipeline.IPipelineGroup) azure.EventConsumer {
	return func(eventData *azeventhubs.ReceivedEventData) error {
		activityLogEventData := new(azure.ActivityLogEventData)
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

func pipelineEventFromRecord(record *azure.ActivityLogEventRecord) *entities.Event {
	rawRecord, err := json.Marshal(record)
	if err != nil {
		return nil
	}

	return &entities.Event{
		PrimaryKeys:   record.PrimaryKeys(),
		OperationType: record.EntityOperationType(),
		Type:          "",
		OriginalRaw:   rawRecord,
	}
}
