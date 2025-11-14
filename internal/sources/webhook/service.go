// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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

func SetupService[T Authentication](
	ctx context.Context,
	router *swagger.Router[fiber.Handler, fiber.Router],
	config Configuration[T],
	p pipeline.IPipelineGroup,
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
		if events.PayloadKey == nil || events.PayloadKey[fiber.MIMEApplicationForm] == "" {
			return nil, fmt.Errorf("%w: FormPayloadKey setting is required for %s", ErrUnsupportedContentType, contentTypeHeader)
		}
		return []byte(c.FormValue(events.PayloadKey[fiber.MIMEApplicationForm])), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedContentType, contentTypeHeader)
	}
}

func webhookHandler[T Authentication](config Configuration[T], p pipeline.IPipelineGroup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := glogrus.FromContext(c.UserContext())

		log.WithFields(map[string]interface{}{
			"sourceType":  "webhook",
			"eventSource": "webhook-event",
			"path":        c.Path(),
			"method":      c.Method(),
		}).Debug("received webhook request")

		if err := config.CheckSignature(c); err != nil {
			log.WithError(err).WithFields(map[string]interface{}{
				"sourceType":  "webhook",
				"eventSource": "webhook-event",
			}).Error("error validating webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		body, err := extractBodyFromContentType(c, config.Events)
		if err != nil {
			log.WithError(err).WithFields(map[string]interface{}{
				"sourceType":  "webhook",
				"eventSource": "webhook-event",
			}).Error("error extracting body from content type")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}
		if len(body) == 0 {
			log.WithFields(map[string]interface{}{
				"sourceType":  "webhook",
				"eventSource": "webhook-event",
			}).Error("empty request body")
			return c.SendStatus(http.StatusOK)
		}

		log.WithFields(map[string]interface{}{
			"sourceType":  "webhook",
			"eventSource": "webhook-event",
			"bodySize":    len(body),
		}).Debug("processing webhook event")

		event, err := config.Events.getPipelineEvent(log, RequestInfo{
			data:    body,
			headers: http.Header(c.GetReqHeaders()),
		})
		if err != nil {
			log.WithError(err).WithFields(map[string]interface{}{
				"sourceType":  "webhook",
				"eventSource": "webhook-event",
			}).Error("error unmarshaling event")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		log.WithFields(map[string]interface{}{
			"sourceType":  "webhook",
			"eventSource": "webhook-event",
			"eventType":   event.GetType(),
			"primaryKeys": event.GetPrimaryKeys().Map(),
			"operation":   event.Operation(),
		}).Debug("webhook event processed successfully, adding to pipeline")

		p.AddMessage(event)

		log.WithFields(map[string]interface{}{
			"sourceType":  "webhook",
			"eventSource": "webhook-event",
			"eventType":   event.GetType(),
		}).Debug("webhook event added to pipeline")

		return c.SendStatus(http.StatusOK)
	}
}
