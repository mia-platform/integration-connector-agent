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

package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/services/lambda"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/services/s3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	aws "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus"
)

type Processor struct {
	logger *logrus.Logger

	config processorConfig
}

type processorConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) entities.Processor {
	return &Processor{
		logger: logger,
		config: processorConfig{
			AccessKeyID:     authOptions.AccessKeyID,
			SecretAccessKey: authOptions.SecretAccessKey.String(),
			SessionToken:    authOptions.SessionToken.String(),
			Region:          authOptions.Region,
		},
	}
}

func (p *Processor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	if input.Operation() == entities.Delete {
		p.logger.Debug("Delete operation detected, skipping event processing")
		return input.Clone(), nil
	}

	cloudTrailEvent, err := p.parseEvent(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	dataProcessor, err := p.EventDataProcessor(cloudTrailEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get event data processor")
		return nil, fmt.Errorf("failed to get event data processor: %w", err)
	}

	newData, err := dataProcessor.GetData(context.Background(), cloudTrailEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get data from AWS service")
		return nil, fmt.Errorf("failed to get data from AWS service: %w", err)
	}

	output := input.Clone()
	output.WithData(newData)
	return output, nil
}

func (p *Processor) EventDataProcessor(cloudTrailEvent awssqsevents.IEvent) (commons.DataAdapter[awssqsevents.IEvent], error) {
	client, err := aws.New(context.Background(), p.logger, aws.Config{
		AccessKeyID:     p.config.AccessKeyID,
		SecretAccessKey: p.config.SecretAccessKey,
		SessionToken:    p.config.SessionToken,
		Region:          cloudTrailEvent.GetRegion(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	eventSource := cloudTrailEvent.EventSource()
	switch eventSource {
	case aws.S3EventSource:
		return s3.New(p.logger, client), nil
	case aws.LambdaEventSource:
		return lambda.New(p.logger, client), nil
	default:
		return nil, fmt.Errorf("unsupported event source: %s", eventSource)
	}
}

func (p *Processor) parseEvent(input entities.PipelineEvent) (awssqsevents.IEvent, error) {
	eventType := input.GetType()
	switch eventType {
	case awssqsevents.RealtimeSyncEventType:
		cloudTrailEvent := new(awssqsevents.CloudTrailEvent)
		if err := json.Unmarshal(input.Data(), &cloudTrailEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
		}
		return cloudTrailEvent, nil
	case awssqsevents.ImportEventType:
		importEvent := new(awssqsevents.CloudTrailImportEvent)
		if err := json.Unmarshal(input.Data(), &importEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
		}
		return importEvent, nil
	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}
