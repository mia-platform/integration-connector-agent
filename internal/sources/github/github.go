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
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"
)

type Config struct {
	WebhookPath    string       `json:"webhookPath"`
	Authentication webhook.HMAC `json:"authentication"`
}

func (c *Config) Validate() error {
	if c.WebhookPath == "" {
		return webhook.ErrWebhookPathRequired
	}
	return nil
}

// getWebhookConfig returns the webhook.Configuration for this source
func (c *Config) getWebhookConfig() (*webhook.Configuration, error) {
	return &webhook.Configuration{
		WebhookPath:    c.WebhookPath,
		Authentication: c.Authentication,
		Events:         &SupportedEvents,
		ContentTypeConfig: &webhook.ContentTypeConfig{
			ContentType: "application/x-www-form-urlencoded",
			Field:       "payload",
		},
	}, nil
}

const (
	pullRequestEvent   = "pull_request"
	defaultWebhookPath = "/github/webhook"
	githubEventHeader  = "X-GitHub-Event"
)

var SupportedEvents = webhook.Events{
	Supported: map[string]webhook.Event{
		pullRequestEvent: {
			Operation: entities.Write,
			GetFieldID: func(parsedData gjson.Result) entities.PkFields {
				id := parsedData.Get("pull_request.id").String()
				if id == "" {
					return nil
				}
				return entities.PkFields{{Key: "pull_request.id", Value: id}}
			},
		},
	},
	EventTypeFieldPath: "_github_event_type", // will be injected from header
}

// This file will contain the implementation for the GitHub webhook source.

func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg *pipeline.Group, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	githubConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}
	webhookConfig, err := githubConfig.getWebhookConfig()
	if err != nil {
		return err
	}
	return webhook.SetupService(ctx, router, webhookConfig, pg)
}
