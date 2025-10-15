// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
	"github.com/mia-platform/integration-connector-agent/internal/utils"

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
		originalBody, decodedBody, wasDecoded := utils.TryDecodeBase64Body(input.Data())
		logFields := logrus.Fields{
			"allowedResultTypes": successResultTypes,
			"resultType":         activityLogEvent.ResultType,
			"reason":             "unsupported_result_type",
			"operationName":      activityLogEvent.OperationName,
			"originalBody":       originalBody,
		}
		if wasDecoded {
			logFields["decodedBody"] = decodedBody
			logFields["wasBase64"] = true
		}
		p.logger.WithFields(logFields).Debug("Event discarded for result type")
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
