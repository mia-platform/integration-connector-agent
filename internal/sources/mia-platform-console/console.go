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
