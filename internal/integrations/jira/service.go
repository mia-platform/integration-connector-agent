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

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/utils"
	"github.com/mia-platform/integration-connector-agent/internal/writer"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	"github.com/sirupsen/logrus"
)

const (
	webhookEndpoint = "/jira/webhook"
)

var (
	ErrEmptyConfiguration      = errors.New("empty configuration")
	ErrUnmarshalEvent          = errors.New("error unmarshaling event")
	ErrUnsupportedWebhookEvent = errors.New("unsupported webhook event")
)

func SetupService(
	ctx context.Context,
	logger *logrus.Entry,
	router *swagger.Router[fiber.Handler, fiber.Router],
	config Configuration,
	writer writer.Writer[entities.PipelineEvent],
) error {
	p, err := pipeline.NewPipeline(logger, writer)
	if err != nil {
		return err
	}

	go func(p pipeline.IPipeline[entities.PipelineEvent]) {
		err := p.Start(ctx)
		if err != nil {
			logger.WithError(err).Error("error starting pipeline")
			// TODO: manage error
			panic(err)
		}
	}(p)

	handler := webhookHandler(config, p)
	if _, err := router.AddRoute(http.MethodPost, webhookEndpoint, handler, swagger.Definitions{}); err != nil {
		return err
	}

	return nil
}

func webhookHandler(config Configuration, p pipeline.IPipeline[entities.PipelineEvent]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		if err := ValidateWebhookRequest(c, config.Secret); err != nil {
			log.WithError(err).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		body := bytes.Clone(c.Body())
		if len(body) == 0 {
			log.Error("empty request body")
			return c.SendStatus(http.StatusOK)
		}

		event, err := getPipelineEvent(body, config.EventIDPath)
		if err != nil {
			log.WithError(err).Error("error unmarshaling event")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		p.AddMessage(event)

		return c.SendStatus(http.StatusOK)
	}
}
