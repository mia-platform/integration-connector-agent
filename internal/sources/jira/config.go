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

package jira

import (
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	defaultWebhookPath    = "/jira/webhook"
	defaultAuthHeaderName = "X-Hub-Signature"

	webhookEventPath = "webhookEvent"
)

type Config struct {
	webhook.Configuration
}

func (c *Config) withDefault() *Config {
	if c.Authentication.HeaderName == "" {
		c.Authentication.HeaderName = defaultAuthHeaderName
	}

	if c.WebhookPath == "" {
		c.WebhookPath = defaultWebhookPath
	}

	if c.Events == nil {
		c.Events = &DefaultSupportedEvents
	}

	if c.Events.EventTypeFieldPath == "" {
		c.Events.EventTypeFieldPath = webhookEventPath
	}

	if c.Events.Supported == nil {
		c.Events.Supported = DefaultSupportedEvents.Supported
	}

	return c
}

func (c *Config) Validate() error {
	c.withDefault()

	if err := c.Configuration.Validate(); err != nil {
		return err
	}

	return nil
}
