// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package server

import (
	"context"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	consolecatalog "github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog"
	crudservice "github.com/mia-platform/integration-connector-agent/internal/sinks/crud-service"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/kafka"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/mongo"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	awssqs "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs"
	azureactivitylogeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub"
	azuredevops "github.com/mia-platform/integration-connector-agent/internal/sources/azure-devops"
	"github.com/mia-platform/integration-connector-agent/internal/sources/confluence"
	gcppubsub "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub"
	"github.com/mia-platform/integration-connector-agent/internal/sources/github"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gitlab"
	"github.com/mia-platform/integration-connector-agent/internal/sources/jboss"
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

func (i Integration) Close(ctx context.Context) error {
	if i.PipelineGroup != nil {
		return i.PipelineGroup.Close(ctx)
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

		sinks, err := setupSinks(ctx, log, cfgPipeline.Sinks)
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

func setupSinks(ctx context.Context, log *logrus.Logger, writers config.Sinks) ([]sinks.Sink[entities.PipelineEvent], error) { //nolint: gocyclo
	var w []sinks.Sink[entities.PipelineEvent]
	for _, configuredWriter := range writers {
		switch configuredWriter.Type {
		case sinks.Mongo:
			config, err := config.GetConfig[*mongo.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			mongoWriter, err := mongo.NewMongoDBWriter[entities.PipelineEvent](ctx, config)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			w = append(w, mongoWriter)
		case sinks.CRUDService:
			config, err := config.GetConfig[*crudservice.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			crudServiceWriter, err := crudservice.NewWriter[entities.PipelineEvent](config)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			w = append(w, crudServiceWriter)
		case sinks.ConsoleCatalog:
			config, err := config.GetConfig[*consolecatalog.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			consoleCatalogWriter, err := consolecatalog.NewWriter[entities.PipelineEvent](config, log)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			w = append(w, consoleCatalogWriter)
		case sinks.Kafka:
			config, err := config.GetConfig[*kafka.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			kafkaSink, err := kafka.New[entities.PipelineEvent](config)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
			}
			w = append(w, kafkaSink)
		case sinks.Fake:
			config, err := config.GetConfig[*fakewriter.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errSetupWriter, err)
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

	if err := setupSource(ctx, log, cfgIntegration.Source, pg, oasRouter, integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func setupSource(ctx context.Context, log *logrus.Logger, source config.GenericConfig, pg pipeline.IPipelineGroup, oasRouter *swagger.Router[fiber.Handler, fiber.Router], integration *Integration) error {
	log.WithFields(logrus.Fields{
		"sourceType": source.Type,
	}).Info("initializing source")

	handlers := map[string]func() error{
		sources.Jira:    func() error { return wrapSetupError(jira.AddSourceToRouter(ctx, source, pg, oasRouter)) },
		sources.Console: func() error { return wrapSetupError(console.AddSourceToRouter(ctx, source, pg, oasRouter)) },
		sources.Confluence: func() error {
			s, err := confluence.NewConfluenceSource(ctx, log, source, pg, oasRouter)
			if err != nil {
				return err
			}
			integration.appendCloseableSource(s)
			return nil
		},
		sources.GCPInventoryPubSub: func() error { return setupGCPInventorySource(ctx, log, source, pg, oasRouter, integration) },
		sources.AWSCloudTrailSQS:   func() error { return setupAWSCloudTrailSource(ctx, log, source, pg, oasRouter, integration) },
		sources.AzureActivityLogEventHub: func() error {
			return wrapSetupError(azureactivitylogeventhub.AddSource(ctx, source, pg, log, oasRouter))
		},
		sources.AzureDevOps: func() error {
			return wrapSetupError(azuredevops.AddSourceToRouter(ctx, source, pg, oasRouter))
		},
		sources.Github: func() error {
			s, err := github.NewGitHubSource(ctx, log, source, pg, oasRouter)
			if err != nil {
				return err
			}
			integration.appendCloseableSource(s)
			return nil
		},
		sources.JBoss: func() error {
			s, err := jboss.NewJBossSource(ctx, log, source, pg, oasRouter)
			if err != nil {
				return err
			}
			integration.appendCloseableSource(s)
			return nil
		},
		sources.Gitlab: func() error {
			s, err := gitlab.NewGitLabSource(ctx, log, source, pg, oasRouter)
			if err != nil {
				return err
			}
			integration.appendCloseableSource(s)
			return nil
		},
	}

	handler, exists := handlers[source.Type]
	if !exists {
		log.WithFields(logrus.Fields{
			"sourceType": source.Type,
		}).Error("source type not supported")
		return fmt.Errorf("source type %s not supported", source.Type)
	}

	log.WithFields(logrus.Fields{
		"sourceType": source.Type,
	}).Debug("setting up source handler")

	if err := handler(); err != nil {
		log.WithFields(logrus.Fields{
			"sourceType": source.Type,
		}).WithError(err).Error("failed to setup source")
		return err
	}

	log.WithFields(logrus.Fields{
		"sourceType": source.Type,
	}).Info("source successfully initialized")

	return nil
}

func wrapSetupError(err error) error {
	if err != nil {
		return fmt.Errorf("%w: %w", errSetupSource, err)
	}
	return nil
}

func setupGCPInventorySource(ctx context.Context, log *logrus.Logger, source config.GenericConfig, pg pipeline.IPipelineGroup, oasRouter *swagger.Router[fiber.Handler, fiber.Router], integration *Integration) error {
	gcpSource, err := gcppubsub.NewInventorySource(ctx, log, source, pg, oasRouter)
	if err != nil {
		return fmt.Errorf("%w: %w", errSetupSource, err)
	}
	integration.appendCloseableSource(gcpSource)
	return nil
}

func setupAWSCloudTrailSource(ctx context.Context, log *logrus.Logger, source config.GenericConfig, pg pipeline.IPipelineGroup, oasRouter *swagger.Router[fiber.Handler, fiber.Router], integration *Integration) error {
	awsConsumer, err := awssqs.NewCloudTrailSource(ctx, log, source, pg, oasRouter)
	if err != nil {
		return fmt.Errorf("%w: %w", errSetupSource, err)
	}
	integration.appendCloseableSource(awsConsumer)
	return nil
}
