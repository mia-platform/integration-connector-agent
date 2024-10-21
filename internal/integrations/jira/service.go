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
	"context"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
)

const (
	webhookEndpoint = "/jira/webhook"
)

func SetupService(ctx context.Context, configPath string, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	config, err := ReadConfiguration(configPath)
	if err != nil {
		return err
	}

	messageChan := make(chan []byte)

	go func() {
		consumeWebhooksData(ctx, messageChan)
	}()

	handler := webhookHandler(config.WebhookSecret(), messageChan)
	if _, err := router.AddRoute(http.MethodPost, webhookEndpoint, handler, swagger.Definitions{}); err != nil {
		return err
	}

	return nil
}

func webhookHandler(secret string, messageChan chan<- []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := ValidateWebhookRequest(c, secret); err != nil {
			// return http error correctly
			return err
		}

		body := []byte{}
		copy(body, c.Body())
		messageChan <- body

		// return 200 ok
		return nil
	}
}

func consumeWebhooksData(ctx context.Context, messageChan chan []byte) {
loop:
	for {
		select {
		case _, open := <-messageChan:
			if !open {
				// the chanel has been closed, break the loop
				break loop
			}
		case <-ctx.Done():
			// context has been cancelled close che channel
			close(messageChan)
		}
	}
}
