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

package confluence

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

const (
	defaultWebhookPath = "/confluence/webhook"
	authHeaderName     = "X-Hub-Signature-256"
)

type Config struct {
	webhook.Configuration[hmac.Authentication]

	// Confluence API configuration
	Username config.SecretSource `json:"username"`
	APIToken config.SecretSource `json:"apiToken"`
	BaseURL  string              `json:"baseUrl"`

	// Item types to import (e.g., "space", "page")
	ItemTypes []string `json:"itemTypes"`

	// Space keys filter for import (optional - filters by specific space keys)
	SpaceKeysFilter []string `json:"spaceKeysFilter"`

	// Import webhook configuration
	ImportWebhookPath    string              `json:"importWebhookPath,omitempty"`
	ImportAuthentication hmac.Authentication `json:"importAuthentication,omitempty"`
}

type ConfluenceSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *Config
	pipeline pipeline.IPipelineGroup
	router   *swagger.Router[fiber.Handler, fiber.Router]
	client   *ConfluenceClient
}

func (c *Config) Validate() error {
	c.withDefault()
	if err := c.Configuration.Validate(); err != nil {
		return err
	}

	// Validate import webhook authentication if import webhook is configured
	if c.ImportWebhookPath != "" {
		if err := c.ImportAuthentication.Validate(); err != nil {
			return err
		}

		// Validate required fields for import functionality
		if c.Username.String() == "" {
			return errors.New("confluence username is required for import functionality")
		}
		if c.APIToken.String() == "" {
			return errors.New("confluence API token is required for import functionality")
		}
		if c.BaseURL == "" {
			return errors.New("confluence baseUrl is required for import functionality")
		}
	}

	return nil
}

//nolint:unparam // Configuration builder pattern, return value is useful API design
func (c *Config) withDefault() *Config {
	c.WebhookPath = cmp.Or(c.WebhookPath, defaultWebhookPath)
	c.Authentication.HeaderName = cmp.Or(c.Authentication.HeaderName, authHeaderName)
	c.Events = cmp.Or(c.Events, SupportedEvents)
	return c
}

func NewConfluenceSource(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.GenericConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) (sources.CloseableSource, error) {
	config, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	var client *ConfluenceClient
	if config.ImportWebhookPath != "" {
		client, err = NewConfluenceClient(config.Username.String(), config.APIToken.String(), config.BaseURL, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create Confluence client: %w", err)
		}
	}

	s := &ConfluenceSource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		router:   oasRouter,
		client:   client,
	}

	if err := s.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize Confluence source: %w", err)
	}

	return s, nil
}

func (s *ConfluenceSource) init() error {
	s.pipeline.Start(s.ctx)

	// Setup webhook endpoint
	if err := webhook.SetupService(s.ctx, s.router, s.config.Configuration, s.pipeline); err != nil {
		return fmt.Errorf("failed to setup webhook service: %w", err)
	}

	// Setup import webhook if configured
	if s.config.ImportWebhookPath != "" {
		s.log.WithField("importWebhookPath", s.config.ImportWebhookPath).Info("Registering Confluence import webhook")
		if err := s.registerImportWebhook(); err != nil {
			return fmt.Errorf("failed to register import webhook: %w", err)
		}
	}

	return nil
}

func (s *ConfluenceSource) Close() error {
	// Confluence client doesn't need explicit closing
	return nil
}

func (s *ConfluenceSource) registerImportWebhook() error {
	apiPath := s.config.ImportWebhookPath
	_, err := s.router.AddRoute(http.MethodPost, apiPath, s.importWebhookHandler, swagger.Definitions{})
	return err
}

// isItemTypeEnabled checks if a specific item type should be imported
func (s *ConfluenceSource) isItemTypeEnabled(itemType string) bool {
	// If no item types are specified, import all types
	if len(s.config.ItemTypes) == 0 {
		return true
	}

	// Check if the item type is in the list
	for _, enabled := range s.config.ItemTypes {
		if enabled == itemType {
			return true
		}
	}
	return false
}

func (s *ConfluenceSource) importWebhookHandler(c *fiber.Ctx) error {
	if err := s.config.ImportAuthentication.CheckSignature(c); err != nil {
		s.log.WithError(err).Error("error validating import webhook request")
		return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "confluence",
		"eventSource": "import-webhook",
		"operation":   "full-import",
		"itemTypes":   s.config.ItemTypes,
	}).Info("starting Confluence full import via webhook")

	// Import spaces (workspaces) if enabled
	if s.isItemTypeEnabled("space") {
		if err := s.importSpaces(c.UserContext()); err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{
				"sourceType":  "confluence",
				"eventSource": "import-webhook",
				"operation":   "import-spaces",
			}).Error("failed to import spaces")
			return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import spaces: " + err.Error()))
		}
	} else {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "confluence",
			"eventSource": "import-webhook",
			"operation":   "import-spaces",
		}).Info("skipping space import - not enabled in itemTypes")
	}

	// Import pages if enabled
	if s.isItemTypeEnabled("page") {
		if err := s.importPages(c.UserContext()); err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{
				"sourceType":  "confluence",
				"eventSource": "import-webhook",
				"operation":   "import-pages",
			}).Error("failed to import pages")
			return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import pages: " + err.Error()))
		}
	} else {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "confluence",
			"eventSource": "import-webhook",
			"operation":   "import-pages",
		}).Info("skipping page import - not enabled in itemTypes")
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "confluence",
		"eventSource": "import-webhook",
		"operation":   "full-import",
	}).Info("Confluence full import completed successfully")
	return c.Status(http.StatusNoContent).SendString("")
}

func (s *ConfluenceSource) importSpaces(ctx context.Context) error {
	importStartTime := time.Now()

	s.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"eventSource":     "import",
		"operation":       "import-spaces",
		"spaceKeysFilter": s.config.SpaceKeysFilter,
	}).Info("starting space import operation")

	s.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"eventSource":     "import",
		"operation":       "list-spaces",
		"spaceKeysFilter": s.config.SpaceKeysFilter,
		"baseURL":         s.config.BaseURL,
	}).Info("about to call Confluence API to list spaces...")

	// Time the API call
	apiCallStartTime := time.Now()
	var spaces []Space
	var err error
	if len(s.config.SpaceKeysFilter) > 0 {
		spaces, err = s.client.ListSpacesWithKeysFilter(ctx, s.config.SpaceKeysFilter)
	} else {
		spaces, err = s.client.ListSpaces(ctx)
	}
	apiCallDuration := time.Since(apiCallStartTime)

	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":      "confluence",
			"eventSource":     "import",
			"operation":       "list-spaces",
			"apiCallDuration": apiCallDuration.String(),
			"totalDuration":   time.Since(importStartTime).String(),
		}).WithError(err).Error("failed to list spaces from Confluence API")
		return fmt.Errorf("failed to list spaces: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"eventSource":     "import",
		"operation":       "list-spaces",
		"spaceCount":      len(spaces),
		"spaceKeysFilter": s.config.SpaceKeysFilter,
		"apiCallDuration": apiCallDuration.String(),
	}).Info("successfully received spaces from Confluence API")

	// Time the event processing
	eventProcessingStartTime := time.Now()
	s.log.WithFields(logrus.Fields{
		"sourceType":  "confluence",
		"eventSource": "import",
		"operation":   "process-spaces",
		"spaceCount":  len(spaces),
	}).Info("starting to process spaces into import events")

	for i, space := range spaces {
		spaceProcessStartTime := time.Now()

		s.log.WithFields(logrus.Fields{
			"sourceType":  "confluence",
			"eventSource": "import",
			"operation":   "import-space",
			"spaceName":   space.Name,
			"spaceKey":    space.Key,
			"spaceId":     space.ID,
			"progress":    fmt.Sprintf("%d/%d", i+1, len(spaces)),
		}).Debug("processing space into import event")

		importEvent := ConfluenceImportEvent{
			Type:    "space",
			ID:      space.ID,
			Key:     space.Key,
			Name:    space.Name,
			BaseURL: s.config.BaseURL,
			Data:    space,
		}

		if err := s.sendImportEvent(importEvent); err != nil {
			s.log.WithFields(logrus.Fields{
				"sourceType":           "confluence",
				"eventSource":          "import",
				"spaceName":            space.Name,
				"spaceKey":             space.Key,
				"spaceProcessDuration": time.Since(spaceProcessStartTime).String(),
			}).WithError(err).Warn("failed to send space import event")
		} else {
			s.log.WithFields(logrus.Fields{
				"sourceType":           "confluence",
				"eventSource":          "import",
				"spaceName":            space.Name,
				"spaceKey":             space.Key,
				"spaceProcessDuration": time.Since(spaceProcessStartTime).String(),
				"progress":             fmt.Sprintf("%d/%d", i+1, len(spaces)),
			}).Debug("successfully sent space import event")
		}
	}

	eventProcessingDuration := time.Since(eventProcessingStartTime)
	totalDuration := time.Since(importStartTime)

	s.log.WithFields(logrus.Fields{
		"sourceType":              "confluence",
		"eventSource":             "import",
		"operation":               "import-spaces",
		"spaceCount":              len(spaces),
		"apiCallDuration":         apiCallDuration.String(),
		"eventProcessingDuration": eventProcessingDuration.String(),
		"totalDuration":           totalDuration.String(),
		"averageSpaceProcessTime": fmt.Sprintf("%.3fs", eventProcessingDuration.Seconds()/float64(max(1, len(spaces)))),
	}).Info("completed space import operation")

	return nil
}

func (s *ConfluenceSource) importPages(ctx context.Context) error {
	spaces, err := s.client.ListSpaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to list spaces for page import: %w", err)
	}

	for _, space := range spaces {
		pages, err := s.client.ListPages(ctx, space.Key)
		if err != nil {
			s.log.WithField("space", space.Key).WithError(err).Warn("failed to list pages for space")
			continue
		}

		for _, page := range pages {
			importEvent := ConfluenceImportEvent{
				Type:     "page",
				ID:       page.ID,
				Key:      page.ID, // Pages use ID as key
				Name:     page.Title,
				BaseURL:  s.config.BaseURL,
				SpaceKey: space.Key,
				Data:     page,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("page", page.Title).WithError(err).Warn("failed to send page import event")
			}
		}
	}

	return nil
}

func (s *ConfluenceSource) sendImportEvent(importEvent ConfluenceImportEvent) error {
	eventStartTime := time.Now()

	s.log.WithFields(logrus.Fields{
		"sourceType":  "confluence",
		"eventSource": "import",
		"eventType":   importEvent.Type,
		"eventId":     importEvent.ID,
		"eventName":   importEvent.Name,
		"baseUrl":     importEvent.BaseURL,
	}).Debug("starting to process import event")

	// Time the JSON marshaling
	marshalStartTime := time.Now()
	data, err := json.Marshal(importEvent)
	marshalDuration := time.Since(marshalStartTime)

	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":      "confluence",
			"eventSource":     "import",
			"eventType":       importEvent.Type,
			"eventId":         importEvent.ID,
			"marshalDuration": marshalDuration.String(),
		}).WithError(err).Error("failed to marshal import event")
		return fmt.Errorf("failed to marshal import event: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"eventSource":     "import",
		"eventType":       importEvent.Type,
		"eventId":         importEvent.ID,
		"marshalDuration": marshalDuration.String(),
		"dataSizeBytes":   len(data),
	}).Debug("successfully marshaled import event data")

	// Time the event builder creation
	builderStartTime := time.Now()
	eventBuilder := NewConfluenceEventBuilder()
	event, err := eventBuilder.GetPipelineEvent(s.ctx, data)
	builderDuration := time.Since(builderStartTime)

	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":      "confluence",
			"eventSource":     "import",
			"eventType":       importEvent.Type,
			"eventId":         importEvent.ID,
			"builderDuration": builderDuration.String(),
		}).WithError(err).Error("failed to create pipeline event")
		return fmt.Errorf("failed to create pipeline event: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"eventSource":     "import",
		"eventType":       importEvent.Type,
		"eventId":         importEvent.ID,
		"builderDuration": builderDuration.String(),
		"pipelineEvent":   event.GetType(),
		"primaryKeys":     event.GetPrimaryKeys().Map(),
	}).Debug("successfully created pipeline event")

	// Time the pipeline addition
	pipelineStartTime := time.Now()
	s.pipeline.AddMessage(event)
	pipelineDuration := time.Since(pipelineStartTime)

	totalDuration := time.Since(eventStartTime)

	s.log.WithFields(logrus.Fields{
		"sourceType":       "confluence",
		"eventSource":      "import",
		"eventType":        importEvent.Type,
		"eventId":          importEvent.ID,
		"eventName":        importEvent.Name,
		"marshalDuration":  marshalDuration.String(),
		"builderDuration":  builderDuration.String(),
		"pipelineDuration": pipelineDuration.String(),
		"totalEventTime":   totalDuration.String(),
		"dataSizeBytes":    len(data),
	}).Debug("successfully sent import event to pipeline")

	return nil
}

// AddSourceToRouter maintains backward compatibility with the existing simple webhook setup
func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	confluenceConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	// Use the simple webhook setup for backward compatibility
	return webhook.SetupService(ctx, router, confluenceConfig.Configuration, pg)
}
