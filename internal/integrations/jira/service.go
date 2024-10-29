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
	"bytes"
	"context"
	"errors"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/data-connector-agent/internal/aggregator"
	"github.com/mia-platform/data-connector-agent/internal/httputil"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

const (
	webhookEndpoint = "/jira/webhook"
)

var (
	ErrEmptyConfiguration = errors.New("empty configuration")
)

func SetupService(ctx context.Context, configPath string, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	config, err := ReadConfiguration(configPath)
	if err != nil {
		return err
	}

	return setupWithConfig(ctx, router, config)
}

func setupWithConfig(_ context.Context, router *swagger.Router[fiber.Handler, fiber.Router], config *Configuration) error {
	if config == nil {
		config = &Configuration{}
	}
	// TODO: here instead to use a buffer size it should be used a proper queue
	messageChan := make(chan []byte, 1000000)

	go func() {
		// consumeWebhooksData(ctx, messageChan)
	}()

	handler := webhookHandler(config.Secret, messageChan)
	if _, err := router.AddRoute(http.MethodPost, webhookEndpoint, handler, swagger.Definitions{}); err != nil {
		return err
	}

	return nil
}

func webhookHandler(secret string, messageChan chan<- []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		if err := ValidateWebhookRequest(c, secret); err != nil {
			log.WithError(err).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(httputil.ValidationError(err.Error()))
		}

		body := bytes.Clone(c.Body())
		if len(body) == 0 {
			log.Error("empty request body")
			return c.SendStatus(http.StatusOK)
		}
		messageChan <- body

		return c.SendStatus(http.StatusOK)
	}
}

func consumeWebhooksData[T any](ctx context.Context, messageChan chan []byte, _ aggregator.IPipeline[T]) {
loop:
	for {
		select {
		case _, open := <-messageChan:
			if !open {
				// the chanel has been closed, break the loop
				break loop
			}
			// TODO: add mapper
			// mappedMsg := msg
			// pipeline.Write(msg)
		case <-ctx.Done():
			// context has been cancelled close che channel
			close(messageChan)
		}
	}
}
