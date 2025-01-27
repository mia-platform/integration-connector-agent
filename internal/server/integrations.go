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

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/mongo"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	"github.com/mia-platform/integration-connector-agent/internal/sources/jira"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// TODO: write an integration test to test this setup
func setupPipelines(ctx context.Context, log *logrus.Logger, cfg *config.Configuration, oasRouter *swagger.Router[fiber.Handler, fiber.Router]) error {
	for _, cfgIntegration := range cfg.Integrations {
		var pipelines []pipeline.IPipeline

		for _, cfgPipeline := range cfgIntegration.Pipelines {
			sinks, err := setupSinks(ctx, cfgPipeline.Sinks)
			if err != nil {
				return err
			}
			if len(sinks) != 1 {
				return fmt.Errorf("only 1 writer is supported, now there are %d", len(sinks))
			}
			writer := sinks[0]

			proc, err := processors.New(cfgPipeline.Processors)
			if err != nil {
				return err
			}

			pip, err := pipeline.New(log, proc, writer)
			if err != nil {
				return err
			}

			pipelines = append(pipelines, pip)
		}

		pg := pipeline.NewGroup(log, pipelines...)

		source := cfgIntegration.Source
		switch source.Type {
		case sources.Jira:
			if err := jira.AddSourceToRouter(ctx, source, pg, oasRouter); err != nil {
				return fmt.Errorf("%w: %s", errSetupSource, err)
			}
		case "test":
			// do nothing only for testing
			return nil
		default:
			return errUnsupportedIntegrationType
		}
	}

	return nil
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
