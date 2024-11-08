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
	integration "github.com/mia-platform/integration-connector-agent/internal/integrations"
	"github.com/mia-platform/integration-connector-agent/internal/integrations/jira"
	"github.com/mia-platform/integration-connector-agent/internal/writer"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/writer/fake"
	"github.com/mia-platform/integration-connector-agent/internal/writer/mongo"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// TODO: write an integration test to test this setup
func setupIntegrations(ctx context.Context, log *logrus.Logger, cfg *config.Configuration, oasRouter *swagger.Router[fiber.Handler, fiber.Router]) error {
	for _, cfgIntegration := range cfg.Integrations {
		writers, err := setupWriters(ctx, cfgIntegration.Writers)
		if err != nil {
			return err
		}
		if len(writers) != 1 {
			return fmt.Errorf("only 1 writer is supported, now there are %d for integration %s", len(writers), cfgIntegration.Type)
		}
		writer := writers[0]

		switch cfgIntegration.Type {
		case integration.Jira:
			// TODO: improve management of configuration
			config := jira.Configuration{
				Secret:      cfgIntegration.Authentication.Secret.String(),
				EventIDPath: cfgIntegration.EventIDPath,
			}

			if err := jira.SetupService(ctx, logrus.NewEntry(log), oasRouter, config, writer); err != nil {
				return err
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

func setupWriters(ctx context.Context, writers []config.Writer) ([]writer.Writer[entities.PipelineEvent], error) {
	var w []writer.Writer[entities.PipelineEvent]
	for _, configuredWriter := range writers {
		switch configuredWriter.Type {
		case writer.Mongo:
			config, err := config.WriterConfig[*mongo.Config](configuredWriter)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			mongoWriter, err := mongo.NewMongoDBWriter[entities.PipelineEvent](ctx, config)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", errSetupWriter, err)
			}
			w = append(w, mongoWriter)
		case writer.Fake:
			config, err := config.WriterConfig[*fakewriter.Config](configuredWriter)
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
