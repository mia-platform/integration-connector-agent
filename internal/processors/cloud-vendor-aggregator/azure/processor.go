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
	"errors"
	"fmt"
	"slices"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"

	"github.com/sirupsen/logrus"
)

var (
	ErrUnsupportedEventSource = errors.New("unsupported event source")
)

type Processor struct {
	logger *logrus.Logger
	client azure.ClientInterface
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) (*Processor, error) {
	client, err := azure.NewClient(authOptions.AuthConfig)
	if err != nil {
		return nil, err
	}

	return &Processor{
		logger: logger,
		client: client,
	}, nil
}

func (p *Processor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	output := input.Clone()

	if input.GetType() == azure.EventTypeFromLiveLoad.String() {
		newData, err := p.GetDataFromLiveEvent(input)
		if err != nil {
			p.logger.WithError(err).Error("Failed to get data from Azure service")
			return nil, fmt.Errorf("failed to get data from Azure service: %w", err)
		}
		output.WithData(newData)
		return output, nil
	}

	activityLogEvent := new(azure.ActivityLogEventRecord)
	if err := json.Unmarshal(input.Data(), &activityLogEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
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

	source := azure.EventSourceFromEvent(activityLogEvent)
	if source == "" {
		err := fmt.Errorf("%w: %s", ErrUnsupportedEventSource, activityLogEvent.OperationName)
		p.logger.WithError(err).Error("Failed to process Azure event")
		return nil, fmt.Errorf("failed to process Azure event: %w", err)
	}

	if input.Operation() == entities.Delete {
		p.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	adapter := NewClient(p.client, source)
	newData, err := adapter.GetData(context.Background(), activityLogEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get data from Azure service")
		return nil, fmt.Errorf("failed to get data from Azure service: %w", err)
	}

	output.WithData(newData)
	return output, nil
}

func (p *Processor) GetDataFromLiveEvent(event entities.PipelineEvent) ([]byte, error) {
	liveData := new(azure.GraphLiveData)
	if err := json.Unmarshal(event.Data(), liveData); err != nil {
		return nil, err
	}

	return json.Marshal(
		commons.NewAsset(liveData.Name, liveData.Type, commons.AzureAssetProvider).
			WithLocation(liveData.Location).
			WithTags(liveData.Tags).
			WithRelationships(azure.RelationshipFromID(liveData.ID)).
			WithRawData(event.Data()),
	)
}
