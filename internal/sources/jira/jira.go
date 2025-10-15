// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package jira

import (
	"cmp"
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultWebhookPath = "/jira/webhook"
	authHeaderName     = "X-Hub-Signature"

	webhookEventPath = "webhookEvent"
)

type Config struct {
	webhook.Configuration[hmac.Authentication]
}

func (c *Config) Validate() error {
	c.withDefault()
	return c.Configuration.Validate()
}

func (c *Config) withDefault() *Config {
	c.WebhookPath = cmp.Or(c.WebhookPath, defaultWebhookPath)
	c.Authentication.HeaderName = cmp.Or(c.Authentication.HeaderName, authHeaderName)
	c.Events = cmp.Or(c.Events, SupportedEvents)
	return c
}

func AddSourceToRouter(
	ctx context.Context,
	cfg config.GenericConfig,
	pg pipeline.IPipelineGroup,
	router *swagger.Router[fiber.Handler, fiber.Router],
) error {
	jiraConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	return webhook.SetupService(ctx, router, jiraConfig.Configuration, pg)
}
