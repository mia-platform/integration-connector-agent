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

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/mia-platform/integration-connector-agent/entities"
	azurecommons "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure/services/functions"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure/services/storage"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	azureactivitylogeventhubevents "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub/events"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/sirupsen/logrus"
)

type Processor struct {
	logger      *logrus.Logger
	credentials azcore.TokenCredential
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) (*Processor, error) {
	credentials, err := azureCredentialFromData(authOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials: %w", err)
	}
	return &Processor{
		logger:      logger,
		credentials: credentials,
	}, nil
}

func (p *Processor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	activityLogEvent := new(azureactivitylogeventhubevents.ActivityLogEventRecord)
	if err := json.Unmarshal(input.Data(), &activityLogEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
	}

	output := input.Clone()

	if input.Operation() == entities.Delete {
		p.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	successResultTypes := []string{"Success", "Succeeded"}
	if !slices.Contains(successResultTypes, activityLogEvent.ResultType) {
		p.logger.
			WithFields(logrus.Fields{
				"allowedResultTypes": successResultTypes,
				"resultType":         activityLogEvent.ResultType,
			}).
			Debug("Event discarded for result type")
		return nil, entities.ErrDiscardEvent
	}

	adapter, err := p.EventDataProcessor(activityLogEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to process Function App event")
		return nil, fmt.Errorf("failed to process Function App event: %w", err)
	}

	newData, err := adapter.GetData(context.Background(), activityLogEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get data from Azure service")
		return nil, fmt.Errorf("failed to get data from Azure service: %w", err)
	}

	output.WithData(newData)
	return output, nil
}

func (p *Processor) EventDataProcessor(activityLogEvent *azureactivitylogeventhubevents.ActivityLogEventRecord) (commons.DataAdapter[*azureactivitylogeventhubevents.ActivityLogEventRecord], error) {
	switch {
	case azurecommons.EventIsForSource(activityLogEvent, storage.EventSource):
		return storage.New(azurecommons.NewClient(p.credentials)), nil
	case azurecommons.EventIsForSource(activityLogEvent, functions.EventSource):
		return functions.New(azurecommons.NewClient(p.credentials)), nil
	default:
		return nil, fmt.Errorf("unsupported event source: %s", activityLogEvent.OperationName)
	}
}

func azureCredentialFromData(config config.AuthOptions) (azcore.TokenCredential, error) {
	credentials := make([]azcore.TokenCredential, 0)

	if len(config.TenantID) > 0 && len(config.ClientID) > 0 && len(config.ClientSecret) > 0 {
		secretCredential, err := azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID.String(),
			config.ClientSecret.String(),
			nil, // Options
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create an Azure client secret credential: %w", err)
		}
		credentials = append(credentials, secretCredential)
	}

	defaultCredential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a default Azure credential: %w", err)
	}
	credentials = append(credentials, defaultCredential)

	return azidentity.NewChainedTokenCredential(credentials, nil)
}
