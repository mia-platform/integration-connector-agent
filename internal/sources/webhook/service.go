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

package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

var (
	ErrUnmarshalEvent          = errors.New("error unmarshaling event")
	ErrUnsupportedWebhookEvent = errors.New("unsupported webhook event")
)

func SetupService(
	ctx context.Context,
	router *swagger.Router[fiber.Handler, fiber.Router],
	config *Configuration,
	p *pipeline.Group,
) error {
	if err := config.Validate(); err != nil {
		return err
	}

	p.Start(ctx)

	handler := webhookHandler(config, p)
	if _, err := router.AddRoute(http.MethodPost, config.WebhookPath, handler, swagger.Definitions{}); err != nil {
		return err
	}

	return nil
}

func webhookHandler(config *Configuration, p *pipeline.Group) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		if err := config.CheckSignature(c); err != nil {
			log.WithError(err).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		body := bytes.Clone(c.Body())
		if len(body) == 0 {
			log.Error("empty request body")
			return c.SendStatus(http.StatusOK)
		}

		event, err := config.Events.getPipelineEvent(log, body)
		if err != nil {
			log.WithError(err).Error("error unmarshaling event")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		p.AddMessage(event)

		return c.SendStatus(http.StatusOK)
	}
}
