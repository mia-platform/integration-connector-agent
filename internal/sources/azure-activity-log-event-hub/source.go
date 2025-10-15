// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azureactivitylogeventhub

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	azureeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-event-hub"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	azure.EventHubConfig

	Authentication hmac.Authentication `json:"authentication,omitempty"`
	WebhookPath    string              `json:"webhookPath"`
}

func (c *Config) Validate() error {
	if err := c.EventHubConfig.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *Config) checkSignature(ctx *fiber.Ctx) error {
	return c.Authentication.CheckSignature(ctx)
}

func AddSource(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	config, err := configFromGeneric(cfg, pg, logger)
	if err != nil {
		return err
	}

	pg.Start(ctx)
	go func(ctx context.Context, config azure.EventHubConfig, logger *logrus.Logger) {
		azureeventhub.SetupEventHub(ctx, config, logger)
	}(ctx, config.EventHubConfig, logger)

	if len(config.WebhookPath) > 0 {
		logger.WithField("webhookPath", config.WebhookPath).Info("Registering import webhook")
		client, err := azure.NewGraphClient(config.AuthConfig)
		if err != nil {
			return err
		}
		_, err = router.AddRoute(
			http.MethodPost,
			config.WebhookPath,
			webhookHandler(client, config, pg), //nolint: contextcheck
			swagger.Definitions{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func configFromGeneric(cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger) (*Config, error) {
	sourceCfg, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	sourceCfg.EventConsumer = activityLogConsumer(pg, logger)
	return sourceCfg, nil
}

func activityLogConsumer(pg pipeline.IPipelineGroup, logger *logrus.Logger) azure.EventConsumer {
	return func(eventData *azeventhubs.ReceivedEventData) error {
		activityLogEventData := new(azure.ActivityLogEventData)
		if err := json.Unmarshal(eventData.Body, activityLogEventData); err != nil {
			logger.WithError(err).Error("failed to unmarshal activity log event data")
			return nil
		}

		for _, record := range activityLogEventData.Records {
			if event := azure.EventFromRecord(record); event != nil {
				pg.AddMessage(event)
			}
		}

		return nil
	}
}

func webhookHandler(client azure.GraphClientInterface, config *Config, pg pipeline.IPipelineGroup) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		if err := config.checkSignature(c); err != nil {
			log.WithError(err).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		supportedTypes := []string{
			azure.StorageAccountEventSource,
			azure.WebSitesEventSource,
			azure.ComputeVirtualMachineEventSource,
			azure.ComputeDiskEventSource,
			azure.VirtualNetworkEventSource,
			azure.NetworkInterfaceEventSource,
			azure.NetworkSecurityGroupEventSource,
			azure.NetworkPublicIPAddressEventSource,
		}
		entities, err := client.Resources(c.UserContext(), supportedTypes)
		if err != nil {
			log.WithError(err).Error("failed to fetch Azure resources")
			return c.Status(http.StatusInternalServerError).JSON(utils.ValidationError(err.Error()))
		}

		for _, entity := range entities {
			pg.AddMessage(entity)
		}

		c.Status(http.StatusNoContent)
		return nil
	}
}
