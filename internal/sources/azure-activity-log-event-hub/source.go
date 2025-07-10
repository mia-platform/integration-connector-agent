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
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	azureeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-event-hub"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Config struct {
	azure.EventHubConfig

	ImportTriggerWebhookPath string `json:"importTriggerWebhookPath,omitempty"`
}

func (c *Config) Validate() error {
	if err := c.EventHubConfig.Validate(); err != nil {
		return err
	}

	return nil
}

func AddSource(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	config, err := configFromGeneric(cfg, pg)
	if err != nil {
		return err
	}

	pg.Start(ctx)
	go func(ctx context.Context, config azure.EventHubConfig, logger *logrus.Logger) {
		azureeventhub.SetupEventHub(ctx, config, logger)
	}(ctx, config.EventHubConfig, logger)

	if len(config.ImportTriggerWebhookPath) > 0 {
		client, err := azure.NewClient(config.AuthConfig)
		if err != nil {
			return err
		}
		_, err = router.AddRoute(
			http.MethodPost,
			config.ImportTriggerWebhookPath,
			webhookHandler(client, pg),
			swagger.Definitions{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func configFromGeneric(cfg config.GenericConfig, pg pipeline.IPipelineGroup) (*Config, error) {
	sourceCfg, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	sourceCfg.EventConsumer = activityLogConsumer(pg)
	return sourceCfg, nil
}

func activityLogConsumer(pg pipeline.IPipelineGroup) azure.EventConsumer {
	return func(eventData *azeventhubs.ReceivedEventData) error {
		activityLogEventData := new(azure.ActivityLogEventData)
		if err := json.Unmarshal(eventData.Body, activityLogEventData); err != nil {
			return fmt.Errorf("failed to read activity log event data: %w", err)
		}

		for _, record := range activityLogEventData.Records {
			if event := azure.EventFromRecord(record); event != nil {
				pg.AddMessage(event)
			}
		}

		return nil
	}
}

func webhookHandler(client azure.ClientInterface, pg pipeline.IPipelineGroup) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}
}
