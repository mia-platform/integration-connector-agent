// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azuredevops

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/basic"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

var (
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidHost          = errors.New("invalid host URL")
)

const (
	defaultAzureWebhookPath = "/azure-devops/webhook"
)

type Config struct {
	webhook.Configuration[*basic.Authentication]

	AzureDevOpsOrganizationURL     string              `json:"azureDevOpsOrganizationUrl"`
	AzureDevOpsPersonalAccessToken config.SecretSource `json:"azureDevOpsPersonalAccessToken"`
	ImportWebhookPath              string              `json:"importWebhookPath"`
	WebhookHost                    string              `json:"webhookHost"`
}

func (c *Config) Validate() error {
	// No specific validation needed for Azure DevOps configuration
	if c.AzureDevOpsOrganizationURL == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredField, "azureDevOpsOrganizationUrl")
	}
	if c.AzureDevOpsPersonalAccessToken.String() == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredField, "azureDevOpsPersonalAccessToken")
	}
	if c.WebhookHost == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredField, "webhookHost")
	} else if _, err := url.Parse(c.WebhookHost); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHost, err)
	}

	c.WebhookPath = cmp.Or(c.WebhookPath, defaultAzureWebhookPath)
	c.Events = supportedEvents
	return c.Configuration.Validate()
}

func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	devopsConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	connection := azuredevops.NewPatConnection(devopsConfig.AzureDevOpsOrganizationURL, devopsConfig.AzureDevOpsPersonalAccessToken.String())

	pg.Start(ctx)
	if err := setupSubscriptions(ctx, connection, devopsConfig); err != nil {
		return fmt.Errorf("failed to setup Azure DevOps source: %w", err)
	}

	if err := webhook.SetupService(ctx, router, devopsConfig.Configuration, pg); err != nil {
		return err
	}

	if devopsConfig.ImportWebhookPath != "" {
		_, err = router.AddRoute(http.MethodPost, devopsConfig.ImportWebhookPath, importWebhookHandler(connection, pg), swagger.Definitions{}) //nolint: contextcheck
		if err != nil {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}

	return nil
}

func importWebhookHandler(connection *azuredevops.Connection, pg pipeline.IPipelineGroup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		repoClient, err := git.NewClient(c.UserContext(), connection)
		if err != nil {
			log.WithError(err).Error("failed to create Azure DevOps git client")
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ValidationError(err.Error()))
		}

		repositories, err := repoClient.GetRepositories(c.UserContext(), git.GetRepositoriesArgs{})
		if err != nil {
			log.WithError(err).Error("failed to get repositories")
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ValidationError(err.Error()))
		}

		for _, repo := range *repositories {
			data, err := json.Marshal(repo)
			if err != nil {
				log.WithError(err).Error("failed to marshal repository data")
				continue
			}

			pg.AddMessage(&entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "repositoryId",
						Value: repo.Id.String(),
					},
					{
						Key:   "type",
						Value: "repository",
					},
				},
				Type:          "azure-devops-repository",
				OperationType: entities.Write,
				OriginalRaw:   data,
			})
		}

		c.Status(http.StatusNoContent)
		return nil
	}
}
