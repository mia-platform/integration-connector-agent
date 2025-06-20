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
	"fmt"
	"mime"
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

var (
	ErrUnmarshalEvent           = errors.New("error unmarshaling event")
	ErrUnsupportedWebhookEvent  = errors.New("unsupported webhook event")
	ErrUnsupportedContentType   = errors.New("unsupported content type for webhook request")
	ErrFailedToParseContentType = errors.New("failed to parse content type")
	ErrFailsToParseBody         = errors.New("failed to parse body")
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

func extractBodyFromContentType(c *fiber.Ctx, events *Events) ([]byte, error) {
	contentTypeHeader := c.Get("Content-Type")

	if contentTypeHeader == "" {
		return bytes.Clone(c.Body()), nil
	}

	mediaType, _, err := mime.ParseMediaType(contentTypeHeader)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedToParseContentType, contentTypeHeader)
	}

	switch mediaType {
	case fiber.MIMEApplicationJSON:
		return bytes.Clone(c.Body()), nil
	case fiber.MIMEApplicationForm:
		if events.FormPayloadKey == "" {
			return nil, fmt.Errorf("%w: FormPayloadKey setting is required for %s", ErrUnsupportedContentType, contentTypeHeader)
		}
		return []byte(c.FormValue(events.FormPayloadKey)), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedContentType, contentTypeHeader)
	}
}

func webhookHandler(config *Configuration, p *pipeline.Group) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		if err := config.CheckSignature(c); err != nil {
			log.WithError(err).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		body, err := extractBodyFromContentType(c, config.Events)
		if err != nil {
			log.WithError(err).Error("error extracting body from content type")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}
		if len(body) == 0 {
			log.Error("empty request body")
			return c.SendStatus(http.StatusOK)
		}

		event, err := config.Events.getPipelineEvent(log, RequestInfo{
			data:    body,
			headers: http.Header(c.GetReqHeaders()),
		})
		if err != nil {
			log.WithError(err).Error("error unmarshaling event")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		p.AddMessage(event)

		return c.SendStatus(http.StatusOK)
	}
}
