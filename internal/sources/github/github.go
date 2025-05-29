// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// This file implements the GitHub webhook source for integration-connector-agent.
//
// GitHub webhook events are validated using the X-Hub-Signature-256 header and a shared secret.
// The secret should be configured as an environment variable or file and referenced in the source config.
//
// The implementation currently supports only the pull_request event, but is structured to allow easy extension for other events.
//
// Security: The webhook secret is required to validate incoming requests. This prevents spoofed events from unauthorized sources.
//
// To add support for more events, extend the SupportedEvents map in github.go.

package github

import (
	"context"
	"encoding/json"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
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

func AddSourceToRouter(ctx context.Context, cfg json.RawMessage, pg *pipeline.Group, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	var c Config
	if err := json.Unmarshal(cfg, &c); err != nil {
		return err
	}
	webhookConfig, err := c.getWebhookConfig()
	if err != nil {
		return err
	}
	return webhook.SetupService(ctx, router, webhookConfig, pg)
}
