// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package console

import (
	"cmp"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	webhookhmac "github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultWebhookPath = "/console/webhook"
	authHeaderName     = "X-Mia-Signature"

	webhookEventPath = "eventName"
)

type Config struct {
	webhook.Configuration[webhookhmac.Authentication]
}

func (c *Config) Validate() error {
	c.withDefault()
	return c.Configuration.Validate()
}

func (c *Config) withDefault() *Config {
	c.WebhookPath = cmp.Or(c.WebhookPath, defaultWebhookPath)
	c.Authentication.HeaderName = authHeaderName
	c.Authentication.CustomValidator = validateBody
	c.Events = cmp.Or(c.Events, SupportedEvents)
	return c
}

func AddSourceToRouter(
	ctx context.Context,
	cfg config.GenericConfig,
	pg pipeline.IPipelineGroup,
	router *swagger.Router[fiber.Handler, fiber.Router],
) error {
	consoleConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	return webhook.SetupService(ctx, router, consoleConfig.Configuration, pg)
}

// validateBody will generate an hmac encoding of bodyData using secret, and than compare it with the expectedSignature
func validateBody(bodyData []byte, secret, expectedSignature string) bool {
	hasher := sha256.New()
	hasher.Write(bodyData)
	hasher.Write([]byte(secret))
	generatedMAC := hasher.Sum(nil)

	expectedMac, err := hex.DecodeString(expectedSignature)
	if err != nil {
		return false
	}

	return hmac.Equal(generatedMAC, expectedMac)
}
