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

package github

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
	defaultWebhookPath = "/github/webhook"
	authHeaderName     = "X-Hub-Signature-256"
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

func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	githubConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	return webhook.SetupService(ctx, router, githubConfig.Configuration, pg)
}
