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

package server

import (
	"context"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	crudservice "github.com/mia-platform/integration-connector-agent/internal/sinks/crud-service"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/mongo"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	awssqs "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	azureactivitylogeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub"
	gcppubsub "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/github"
	"github.com/mia-platform/integration-connector-agent/internal/sources/jira"
	console "github.com/mia-platform/integration-connector-agent/internal/sources/mia-platform-console"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Integration struct {
	PipelineGroup  pipeline.IPipelineGroup
	sourcesToClose []sources.CloseableSource
}

func (i *Integration) appendCloseableSource(source sources.CloseableSource) {
	if i.sourcesToClose == nil {
		i.sourcesToClose = make([]sources.CloseableSource, 0)
	}
	i.sourcesToClose = append(i.sourcesToClose, source)
}

func (i Integration) Close() error {
	if i.PipelineGroup != nil {
		return i.PipelineGroup.Close()
	}
	for _, source := range i.sourcesToClose {
		if err := source.Close(); err != nil {
			return fmt.Errorf("error closing source: %w", err)
		}
	}
	return nil
}

// TODO: write an integration test to test this setup
func setupIntegrations(ctx context.Context, log *logrus.Logger, cfg *config.Configuration, oasRouter *swagger.Router[fiber.Handler, fiber.Router]) ([]*Integration, error) {
	integrations := make([]*Integration, 0)
	for _, cfgIntegration := range cfg.Integrations {
		log.WithFields(logrus.Fields{
			"sourceType":   cfgIntegration.Source.Type,
			"pipelinesLen": len(cfgIntegration.Pipelines),
		}).Trace("setting up integration")

		pipelines, err := setupIntegrationPipelines(ctx, log, cfgIntegration)
		if err != nil {
			return nil, err
		}

		pg := pipeline.NewGroup(log, pipelines...)

		// skip this source as it is only used for test
		if cfgIntegration.Source.Type == "test" {
			continue
		}

		integration, err := runIntegration(ctx, log, pg, cfgIntegration, oasRouter)
		if err != nil {
			return nil, err
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func setupIntegrationPipelines(ctx context.Context, log *logrus.Logger, cfgIntegration config.Integration) ([]pipeline.IPipeline, error) {
	pipelines := make([]pipeline.IPipeline, 0)

	for i, cfgPipeline := range cfgIntegration.Pipelines {
		log.WithFields(logrus.Fields{
			"sourceType":    cfgIntegration.Source.Type,
			"pipelineIndex": i,
			"processorsLen": len(cfgPipeline.Processors),
		}).Trace("setting up pipeline processors")

		sinks, err := setupSinks(ctx, cfgPipeline.Sinks)
		if err != nil {
			return nil, err
		}
		if len(sinks) != 1 {
			return nil, fmt.Errorf("only 1 writer is supported, now there are %d", len(sinks))
		}
		writer := sinks[0]

		proc, err := processors.New(log, cfgPipeline.Processors)
		if err != nil {
			return nil, err
		}

		pip, err := pipeline.New(log, proc, writer)
		if err != nil {
			return nil, err
		}

		pipelines = append(pipelines, pip)
	}
	return pipelines, nil
}

func setupSinks(ctx context.Context, writers config.Sinks) ([]sinks.Sink[entities.PipelineEvent], error) {
	var w []sinks.Sink[entities.PipelineEvent]
	for _, configuredWriter := range writers {
		switch configuredWriter.Type {
		case sinks.Mongo:
			config, err := config.GetConfig[*mongo.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			mongoWriter, err := mongo.NewMongoDBWriter[entities.PipelineEvent](ctx, config)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			w = append(w, mongoWriter)
		case sinks.CRUDService:
			config, err := config.GetConfig[*crudservice.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			crudServiceWriter, err := crudservice.NewWriter[entities.PipelineEvent](config)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			w = append(w, crudServiceWriter)
		case sinks.Fake:
			config, err := config.GetConfig[*fakewriter.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			w = append(w, fakewriter.New(config))
		default:
			return nil, fmt.Errorf("%w: %s", errUnsupportedWriter, configuredWriter.Type)
		}
	}

	return w, nil
}

func runIntegration(ctx context.Context, log *logrus.Logger, pg pipeline.IPipelineGroup, cfgIntegration config.Integration, oasRouter *swagger.Router[fiber.Handler, fiber.Router]) (*Integration, error) {
	integration := &Integration{
		PipelineGroup: pg,
	}
	source := cfgIntegration.Source
	switch source.Type {
	case sources.Jira:
		if err := jira.AddSourceToRouter(ctx, source, pg, oasRouter); err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}
	case sources.Console:
		if err := console.AddSourceToRouter(ctx, source, pg, oasRouter); err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}
	case sources.GCPInventoryPubSub:
		pubsub, err := gcppubsub.New(&gcppubsub.ConsumerOptions{
			Ctx: ctx,
			Log: log,
		}, source, pg, gcppubsubevents.NewInventoryEventBuilder())
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}

		integration.appendCloseableSource(pubsub)
	case sources.AWSCloudTrailSQS:
		awsConsumer, err := awssqs.New(&awssqs.ConsumerOptions{
			Ctx: ctx,
			Log: log,
		}, source, pg, awssqsevents.NewCloudTrailEventBuilder())
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}
		integration.appendCloseableSource(awsConsumer)
	case sources.AzureActivityLogEventHub:
		if err := azureactivitylogeventhub.AddSource(ctx, source, pg, log); err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}
	case sources.Github:
		if err := github.AddSourceToRouter(ctx, source, pg, oasRouter); err != nil {
			return nil, fmt.Errorf("%w: %s", errSetupSource, err)
		}
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedIntegrationType, source.Type)
	}

	return integration, nil
}
