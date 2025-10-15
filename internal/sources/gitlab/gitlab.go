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

package gitlab

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	defaultWebhookPath = "/gitlab/webhook"
	authHeaderName     = "X-Gitlab-Token"
)

type Config struct {
	webhook.Configuration[hmac.Authentication]

	// GitLab API configuration
	Token   config.SecretSource `json:"token"`
	BaseURL string              `json:"baseUrl,omitempty"`
	Group   string              `json:"group,omitempty"`

	// Import webhook configuration
	ImportWebhookPath    string              `json:"importWebhookPath,omitempty"`
	ImportAuthentication hmac.Authentication `json:"importAuthentication,omitempty"`
}

type GitLabSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *Config
	pipeline pipeline.IPipelineGroup
	router   *swagger.Router[fiber.Handler, fiber.Router]
	client   *GitLabClient
}

func (c *Config) Validate() error {
	c.withDefault()
	if err := c.Configuration.Validate(); err != nil {
		return err
	}

	// Validate import webhook authentication if import webhook is configured
	if c.ImportWebhookPath != "" {
		return c.ImportAuthentication.Validate()
	}

	return nil
}

func (c *Config) withDefault() *Config {
	c.WebhookPath = cmp.Or(c.WebhookPath, defaultWebhookPath)
	c.Authentication.HeaderName = cmp.Or(c.Authentication.HeaderName, authHeaderName)
	c.Events = cmp.Or(c.Events, SupportedEvents)
	c.BaseURL = cmp.Or(c.BaseURL, "https://gitlab.com")
	return c
}

func NewGitLabSource(
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

	var client *GitLabClient
	if config.ImportWebhookPath != "" {
		// Check for GitLab API token
		if config.Token.String() == "" {
			return nil, errors.New("GitLab API token is required for import functionality")
		}

		if config.Group == "" {
			return nil, errors.New("GitLab group is required for import functionality")
		}

		client, err = NewGitLabClient(config.Token.String(), config.BaseURL, config.Group)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab client: %w", err)
		}
	}

	s := &GitLabSource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		router:   oasRouter,
		client:   client,
	}

	if err := s.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GitLab source: %w", err)
	}

	return s, nil
}

func (s *GitLabSource) init() error {
	s.pipeline.Start(s.ctx)

	// Setup webhook endpoint
	if err := webhook.SetupService(s.ctx, s.router, s.config.Configuration, s.pipeline); err != nil {
		return fmt.Errorf("failed to setup webhook service: %w", err)
	}

	// Setup import webhook if configured
	if s.config.ImportWebhookPath != "" {
		s.log.WithField("importWebhookPath", s.config.ImportWebhookPath).Info("Registering GitLab import webhook")
		if err := s.registerImportWebhook(); err != nil {
			return fmt.Errorf("failed to register import webhook: %w", err)
		}
	}

	return nil
}

func (s *GitLabSource) Close() error {
	// GitLab client doesn't need explicit closing
	return nil
}

func (s *GitLabSource) registerImportWebhook() error {
	apiPath := s.config.ImportWebhookPath
	_, err := s.router.AddRoute(http.MethodPost, apiPath, s.importWebhookHandler, swagger.Definitions{})
	return err
}

func (s *GitLabSource) importWebhookHandler(c *fiber.Ctx) error {
	if err := s.config.ImportAuthentication.CheckSignature(c); err != nil {
		s.log.WithError(err).Error("error validating import webhook request")
		return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import-webhook",
		"operation":   "full-import",
	}).Info("starting GitLab full import via webhook")

	// Import projects
	if err := s.importProjects(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "import-projects",
		}).Error("failed to import projects")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import projects: " + err.Error()))
	}

	// Import merge requests
	if err := s.importMergeRequests(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "import-merge-requests",
		}).Error("failed to import merge requests")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import merge requests: " + err.Error()))
	}

	// Import pipelines
	if err := s.importPipelines(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "import-pipelines",
		}).Error("failed to import pipelines")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import pipelines: " + err.Error()))
	}

	// Import releases
	if err := s.importReleases(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "import-releases",
		}).Error("failed to import releases")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import releases: " + err.Error()))
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import-webhook",
		"operation":   "full-import",
	}).Info("GitLab full import completed successfully")
	return c.Status(http.StatusNoContent).SendString("")
}

func (s *GitLabSource) importProjects(ctx context.Context) error {
	s.log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "list-projects",
		"group":       s.config.Group,
	}).Debug("starting project import")

	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":   "gitlab",
		"eventSource":  "import",
		"operation":    "list-projects",
		"group":        s.config.Group,
		"projectCount": len(projects),
	}).Info("found projects for import")

	for i, project := range projects {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import",
			"operation":   "import-project",
			"projectName": project.Name,
			"projectId":   project.ID,
			"progress":    fmt.Sprintf("%d/%d", i+1, len(projects)),
		}).Debug("importing project")

		// Fetch README content for the project
		readmeContent, err := s.client.GetProjectReadme(ctx, project.ID)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"sourceType":  "gitlab",
				"eventSource": "import",
				"projectName": project.Name,
				"projectId":   project.ID,
			}).WithError(err).Warn("failed to fetch README content, using project description as fallback")
			// Use project description as fallback
			readmeContent = project.Description
		} else if readmeContent == "" {
			// If README is empty, use project description as fallback
			readmeContent = project.Description
		}

		// Set the README content in the project struct
		project.ReadmeContent = readmeContent

		importEvent := GitLabImportEvent{
			Type:     "project",
			ID:       project.ID,
			Name:     project.Name,
			FullName: project.PathWithNamespace,
			Group:    s.config.Group,
			Data:     project,
		}

		if err := s.sendImportEvent(importEvent); err != nil {
			s.log.WithFields(logrus.Fields{
				"sourceType":  "gitlab",
				"eventSource": "import",
				"projectName": project.Name,
				"projectId":   project.ID,
			}).WithError(err).Warn("failed to send project import event")
		}
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":   "gitlab",
		"eventSource":  "import",
		"operation":    "import-projects",
		"projectCount": len(projects),
	}).Info("completed project import")

	return nil
}

func (s *GitLabSource) importMergeRequests(ctx context.Context) error {
	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects for MR import: %w", err)
	}

	for _, project := range projects {
		mergeRequests, err := s.client.ListMergeRequests(ctx, project.ID)
		if err != nil {
			s.log.WithField("project", project.Name).WithError(err).Warn("failed to list merge requests for project")
			continue
		}

		for _, mr := range mergeRequests {
			importEvent := GitLabImportEvent{
				Type:      "merge_request",
				ID:        mr.ID,
				Name:      fmt.Sprintf("MR !%d", mr.IID),
				FullName:  fmt.Sprintf("%s!%d", project.PathWithNamespace, mr.IID),
				Group:     s.config.Group,
				ProjectID: project.ID,
				Data:      mr,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("mergeRequest", mr.IID).WithError(err).Warn("failed to send merge request import event")
			}
		}
	}

	return nil
}

func (s *GitLabSource) importPipelines(ctx context.Context) error {
	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects for pipeline import: %w", err)
	}

	for _, project := range projects {
		pipelines, err := s.client.ListPipelines(ctx, project.ID)
		if err != nil {
			s.log.WithField("project", project.Name).WithError(err).Warn("failed to list pipelines for project")
			continue
		}

		for _, pipeline := range pipelines {
			importEvent := GitLabImportEvent{
				Type:      "pipeline",
				ID:        pipeline.ID,
				Name:      fmt.Sprintf("Pipeline #%d", pipeline.ID),
				FullName:  fmt.Sprintf("%s#%d", project.PathWithNamespace, pipeline.ID),
				Group:     s.config.Group,
				ProjectID: project.ID,
				Data:      pipeline,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("pipeline", pipeline.ID).WithError(err).Warn("failed to send pipeline import event")
			}
		}
	}

	return nil
}

func (s *GitLabSource) importReleases(ctx context.Context) error {
	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects for release import: %w", err)
	}

	for _, project := range projects {
		releases, err := s.client.ListReleases(ctx, project.ID)
		if err != nil {
			s.log.WithField("project", project.Name).WithError(err).Warn("failed to list releases for project")
			continue
		}

		for _, release := range releases {
			importEvent := GitLabImportEvent{
				Type:      "release",
				ID:        int64(len(release.TagName)), // Use tag name length as ID since GitLab releases don't have numeric IDs
				Name:      release.Name,
				FullName:  fmt.Sprintf("%s@%s", project.PathWithNamespace, release.TagName),
				Group:     s.config.Group,
				ProjectID: project.ID,
				Data:      release,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("release", release.TagName).WithError(err).Warn("failed to send release import event")
			}
		}
	}

	return nil
}

func (s *GitLabSource) sendImportEvent(importEvent GitLabImportEvent) error {
	s.log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"eventType":   importEvent.Type,
		"eventId":     importEvent.ID,
		"eventName":   importEvent.Name,
		"group":       importEvent.Group,
	}).Debug("sending import event to pipeline")

	data, err := json.Marshal(importEvent)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import",
			"eventType":   importEvent.Type,
			"eventId":     importEvent.ID,
		}).WithError(err).Error("failed to marshal import event")
		return fmt.Errorf("failed to marshal import event: %w", err)
	}

	// Create a mock event builder similar to other sources
	eventBuilder := NewGitLabEventBuilder()
	event, err := eventBuilder.GetPipelineEvent(s.ctx, data)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import",
			"eventType":   importEvent.Type,
			"eventId":     importEvent.ID,
		}).WithError(err).Error("failed to create pipeline event")
		return fmt.Errorf("failed to create pipeline event: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":    "gitlab",
		"eventSource":   "import",
		"eventType":     importEvent.Type,
		"eventId":       importEvent.ID,
		"pipelineEvent": event.GetType(),
		"primaryKeys":   event.GetPrimaryKeys().Map(),
	}).Debug("import event successfully created, adding to pipeline")

	s.pipeline.AddMessage(event)
	return nil
}

// AddSourceToRouter maintains backward compatibility with the existing simple webhook setup
func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	gitlabConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	// Use the simple webhook setup for backward compatibility
	return webhook.SetupService(ctx, router, gitlabConfig.Configuration, pg)
}
