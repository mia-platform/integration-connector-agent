// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gitlab

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/token"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	"github.com/sirupsen/logrus"
)

const (
	defaultWebhookPath = "/gitlab/webhook"
	authHeaderName     = "X-Gitlab-Token"
)

type Config struct {
	webhook.Configuration[token.Authentication]

	// GitLab API configuration
	Token   config.SecretSource `json:"token"`
	BaseURL string              `json:"baseUrl,omitempty"`
	Group   string              `json:"group,omitempty"`

	// Import webhook configuration
	ImportWebhookPath    string              `json:"importWebhookPath,omitempty"`
	ImportAuthentication hmac.Authentication `json:"importAuthentication,omitempty"`
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

		if c.Token.String() == "" {
			return errors.New("GitLab API token is required for import functionality")
		}

		if c.Group == "" {
			return errors.New("GitLab group is required for import functionality")
		}
	}

	return nil
}

func (c *Config) withDefault() *Config {
	c.WebhookPath = cmp.Or(c.WebhookPath, defaultWebhookPath)
	c.Authentication.HeaderName = cmp.Or(c.Authentication.HeaderName, authHeaderName)
	c.BaseURL = cmp.Or(c.BaseURL, "https://gitlab.com")
	c.Events = cmp.Or(c.Events, supportedEvents(c.BaseURL))
	return c
}

func AddSourceToRouter(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, router *swagger.Router[fiber.Handler, fiber.Router]) error {
	gitlabConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	client, err := NewGitLabClient(gitlabConfig.Token.String(), gitlabConfig.BaseURL, gitlabConfig.Group)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	if len(gitlabConfig.ImportWebhookPath) > 0 {
		_, err := router.AddRoute(
			http.MethodPost,
			gitlabConfig.ImportWebhookPath,
			webhookHandlerImport(client, gitlabConfig, pg), //nolint: contextcheck
			swagger.Definitions{})
		if err != nil {
			return err
		}
	}

	// Use the simple webhook setup for backward compatibility
	return webhook.SetupService(ctx, router, gitlabConfig.Configuration, pg)
}

func webhookHandlerImport(client *GitLabClient, cfg *Config, pg pipeline.IPipelineGroup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		log := glogrus.FromContext(ctx)
		if err := cfg.ImportAuthentication.CheckSignature(c); err != nil {
			log.WithError(err).Error("error validating import webhook request")
			return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
		}

		log.WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "full-import",
		}).Info("starting GitLab full import via webhook")

		// Import projects
		projectIDs, err := importProjects(ctx, client, pg, log)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"sourceType":  "gitlab",
				"eventSource": "import-webhook",
				"operation":   "import-projects",
			}).Error("failed to import projects")
			return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import projects: " + err.Error()))
		}

		// Import merge requests
		importMergeRequests(ctx, client, projectIDs, pg, log)

		// Import pipelines
		importPipelines(ctx, client, projectIDs, pg, log)

		// Import releases
		importReleases(ctx, client, projectIDs, pg, log)

		log.WithFields(logrus.Fields{
			"sourceType":  "gitlab",
			"eventSource": "import-webhook",
			"operation":   "full-import",
		}).Info("GitLab full import completed successfully")

		c.Status(http.StatusNoContent)
		return nil
	}
}

func importProjects(ctx context.Context, client *GitLabClient, pg pipeline.IPipelineGroup, log *logrus.Entry) ([]string, error) {
	group := client.group
	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-projects",
		"group":       group,
	}).Debug("starting project import")

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projectIDs := make([]string, 0, len(projects))
	for _, project := range projects {
		projectID := strconv.FormatInt(project["id"].(int64), 10)
		// cannot fail because we know that project is serializable
		data, _ := json.Marshal(project)
		event := &entities.Event{
			PrimaryKeys: entities.PkFields{
				{Key: "id", Value: projectID},
				{Key: "url", Value: client.baseURL.String()},
			},
			Type:          "gitlab:project:import",
			OperationType: entities.Write,
			OriginalRaw:   data,
		}
		pg.AddMessage(event)
		projectIDs = append(projectIDs, projectID)
	}

	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-projects",
		"group":       group,
	}).Debug("end project import")
	return projectIDs, nil
}

func importMergeRequests(ctx context.Context, client *GitLabClient, projectIDs []string, pg pipeline.IPipelineGroup, log *logrus.Entry) {
	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-mrs",
	}).Debug("starting mr import")
	for _, projectID := range projectIDs {
		mergeRequests, err := client.ListMergeRequests(ctx, projectID)
		if err != nil {
			log.WithField("project", projectID).WithError(err).Warn("failed to list merge requests for project")
			continue
		}

		for _, mr := range mergeRequests {
			mrID := mr["id"].(string)
			// cannot fail because we know that mr is serializable
			data, _ := json.Marshal(mr)
			event := &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "projectId", Value: projectID},
					{Key: "id", Value: mrID},
					{Key: "url", Value: client.baseURL.String()},
				},
				Type:          "gitlab:merge_request:import",
				OperationType: entities.Write,
				OriginalRaw:   data,
			}
			pg.AddMessage(event)
		}
	}

	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-mrs",
	}).Debug("end mr import")
}

func importPipelines(ctx context.Context, client *GitLabClient, projectIDs []string, pg pipeline.IPipelineGroup, log *logrus.Entry) {
	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-pipelines",
	}).Debug("starting pipelines import")

	for _, projectID := range projectIDs {
		pipelines, err := client.ListPipelines(ctx, projectID)
		if err != nil {
			log.WithField("project", projectID).WithError(err).Warn("failed to list pipelines for project")
			continue
		}

		for _, pipeline := range pipelines {
			pipelineID := pipeline["id"].(string)
			// cannot fail because we know that mr is serializable
			data, _ := json.Marshal(pipeline)
			event := &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "projectId", Value: projectID},
					{Key: "id", Value: pipelineID},
					{Key: "url", Value: client.baseURL.String()},
				},
				Type:          "gitlab:pipeline:import",
				OperationType: entities.Write,
				OriginalRaw:   data,
			}
			pg.AddMessage(event)
		}
	}

	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-pipelines",
	}).Debug("end pipelines import")
}

func importReleases(ctx context.Context, client *GitLabClient, projectIDs []string, pg pipeline.IPipelineGroup, log *logrus.Entry) {
	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-releases",
	}).Debug("starting releases import")

	for _, projectID := range projectIDs {
		releases, err := client.ListReleases(ctx, projectID)
		if err != nil {
			log.WithField("project", projectID).WithError(err).Warn("failed to list releases for project")
			continue
		}

		for _, release := range releases {
			// cannot fail because we know that mr is serializable
			data, _ := json.Marshal(release)
			event := &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "projectId", Value: projectID},
					{Key: "tagName", Value: release["tag_name"].(string)},
					{Key: "url", Value: client.baseURL.String()},
				},
				Type:          "gitlab:relase:import",
				OperationType: entities.Write,
				OriginalRaw:   data,
			}
			pg.AddMessage(event)
		}
	}

	log.WithFields(logrus.Fields{
		"sourceType":  "gitlab",
		"eventSource": "import",
		"operation":   "import-releases",
	}).Debug("end releases import")
}
