// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package github

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
	defaultWebhookPath = "/github/webhook"
	authHeaderName     = "X-Hub-Signature-256"
)

type Config struct {
	webhook.Configuration[hmac.Authentication]

	// GitHub API configuration
	ClientID     config.SecretSource `json:"clientId"`
	ClientSecret config.SecretSource `json:"clientSecret"`
	Organization string              `json:"organization,omitempty"`

	// Legacy token support (deprecated)
	Token config.SecretSource `json:"token,omitempty"`

	// Import webhook configuration
	ImportWebhookPath    string              `json:"importWebhookPath,omitempty"`
	ImportAuthentication hmac.Authentication `json:"importAuthentication,omitempty"`
}

type GitHubSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *Config
	pipeline pipeline.IPipelineGroup
	router   *swagger.Router[fiber.Handler, fiber.Router]
	client   *GitHubClient
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
	return c
}

func NewGitHubSource(
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

	var client *GitHubClient
	if config.ImportWebhookPath != "" {
		// Check for GitHub authentication type
		switch {
		case config.ClientID.String() != "" && config.ClientSecret.String() != "":
			// GitHub App authentication (preferred)
			if config.Organization == "" {
				return nil, errors.New("GitHub organization is required for import functionality")
			}

			client, err = NewGitHubClientWithApp(config.ClientID.String(), config.ClientSecret.String(), config.Organization)
			if err != nil {
				return nil, fmt.Errorf("failed to create GitHub client with App authentication: %w", err)
			}
		case config.Token.String() != "":
			// Fallback to legacy token authentication
			if config.Organization == "" {
				return nil, errors.New("GitHub organization is required for import functionality")
			}

			client, err = NewGitHubClient(config.Token.String(), config.Organization)
			if err != nil {
				return nil, fmt.Errorf("failed to create GitHub client: %w", err)
			}
		default:
			return nil, errors.New("GitHub authentication is required for import functionality: either clientId/clientSecret or token must be provided")
		}
	}

	s := &GitHubSource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		router:   oasRouter,
		client:   client,
	}

	if err := s.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub source: %w", err)
	}

	return s, nil
}

func (s *GitHubSource) init() error {
	s.pipeline.Start(s.ctx)

	// Setup webhook endpoint
	if err := webhook.SetupService(s.ctx, s.router, s.config.Configuration, s.pipeline); err != nil {
		return fmt.Errorf("failed to setup webhook service: %w", err)
	}

	// Setup import webhook if configured
	if s.config.ImportWebhookPath != "" {
		s.log.WithField("importWebhookPath", s.config.ImportWebhookPath).Info("Registering GitHub import webhook")
		if err := s.registerImportWebhook(); err != nil {
			return fmt.Errorf("failed to register import webhook: %w", err)
		}
	}

	return nil
}

func (s *GitHubSource) Close() error {
	// GitHub client doesn't need explicit closing
	return nil
}

func (s *GitHubSource) registerImportWebhook() error {
	apiPath := s.config.ImportWebhookPath
	_, err := s.router.AddRoute(http.MethodPost, apiPath, s.importWebhookHandler, swagger.Definitions{})
	return err
}

func (s *GitHubSource) importWebhookHandler(c *fiber.Ctx) error {
	if err := s.config.ImportAuthentication.CheckSignature(c); err != nil {
		s.log.WithError(err).Error("error validating import webhook request")
		return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "github",
		"eventSource": "import-webhook",
		"operation":   "full-import",
	}).Info("starting GitHub full import via webhook")

	// Import repositories
	if err := s.importRepositories(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import-webhook",
			"operation":   "import-repositories",
		}).Error("failed to import repositories")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import repositories: " + err.Error()))
	}

	// Import pull requests
	if err := s.importPullRequests(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import-webhook",
			"operation":   "import-pull-requests",
		}).Error("failed to import pull requests")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import pull requests: " + err.Error()))
	}

	// Import workflow runs
	if err := s.importWorkflowRuns(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import-webhook",
			"operation":   "import-workflow-runs",
		}).Error("failed to import workflow runs")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import workflow runs: " + err.Error()))
	}

	// Import issues
	if err := s.importIssues(c.UserContext()); err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import-webhook",
			"operation":   "import-issues",
		}).Error("failed to import issues")
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to import issues: " + err.Error()))
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":  "github",
		"eventSource": "import-webhook",
		"operation":   "full-import",
	}).Info("GitHub full import completed successfully")
	return c.Status(http.StatusNoContent).SendString("")
}

func (s *GitHubSource) importRepositories(ctx context.Context) error {
	s.log.WithFields(logrus.Fields{
		"sourceType":   "github",
		"eventSource":  "import",
		"operation":    "list-repositories",
		"organization": s.config.Organization,
	}).Debug("starting repository import")

	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":      "github",
		"eventSource":     "import",
		"operation":       "list-repositories",
		"organization":    s.config.Organization,
		"repositoryCount": len(repositories),
	}).Info("found repositories for import")

	for i, repo := range repositories {
		s.log.WithFields(logrus.Fields{
			"sourceType":     "github",
			"eventSource":    "import",
			"operation":      "import-repository",
			"repositoryName": repo.Name,
			"repositoryId":   repo.ID,
			"progress":       fmt.Sprintf("%d/%d", i+1, len(repositories)),
		}).Debug("importing repository")

		// Fetch README content for the repository
		readmeContent, err := s.client.GetRepositoryReadme(ctx, repo.Name)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"sourceType":     "github",
				"eventSource":    "import",
				"repositoryName": repo.Name,
				"repositoryId":   repo.ID,
			}).WithError(err).Warn("failed to fetch README content, using repository description as fallback")
			// Use repository description as fallback
			readmeContent = repo.Description
		} else if readmeContent == "" {
			// If README is empty, use repository description as fallback
			readmeContent = repo.Description
		}

		// Set the README content in the repository struct
		repo.ReadmeContent = readmeContent

		importEvent := GitHubImportEvent{
			Type:         "repository",
			ID:           repo.ID,
			Name:         repo.Name,
			FullName:     repo.FullName,
			Organization: s.config.Organization,
			Data:         repo,
		}

		if err := s.sendImportEvent(importEvent); err != nil {
			s.log.WithFields(logrus.Fields{
				"sourceType":     "github",
				"eventSource":    "import",
				"repositoryName": repo.Name,
				"repositoryId":   repo.ID,
			}).WithError(err).Warn("failed to send repository import event")
		}
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":      "github",
		"eventSource":     "import",
		"operation":       "import-repositories",
		"repositoryCount": len(repositories),
	}).Info("completed repository import")

	return nil
}

func (s *GitHubSource) importPullRequests(ctx context.Context) error {
	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories for PR import: %w", err)
	}

	for _, repo := range repositories {
		pullRequests, err := s.client.ListPullRequests(ctx, repo.Name)
		if err != nil {
			s.log.WithField("repository", repo.Name).WithError(err).Warn("failed to list pull requests for repository")
			continue
		}

		for _, pr := range pullRequests {
			importEvent := GitHubImportEvent{
				Type:         "pull_request",
				ID:           pr.ID,
				Name:         fmt.Sprintf("PR #%d", pr.Number),
				FullName:     fmt.Sprintf("%s#%d", repo.FullName, pr.Number),
				Organization: s.config.Organization,
				Repository:   repo.Name,
				Data:         pr,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("pullRequest", pr.Number).WithError(err).Warn("failed to send pull request import event")
			}
		}
	}

	return nil
}

func (s *GitHubSource) importWorkflowRuns(ctx context.Context) error {
	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories for workflow runs import: %w", err)
	}

	for _, repo := range repositories {
		workflowRuns, err := s.client.ListWorkflowRuns(ctx, repo.Name)
		if err != nil {
			s.log.WithField("repository", repo.Name).WithError(err).Warn("failed to list workflow runs for repository")
			continue
		}

		for _, run := range workflowRuns {
			importEvent := GitHubImportEvent{
				Type:         "workflow_run",
				ID:           run.ID,
				Name:         run.Name,
				FullName:     fmt.Sprintf("%s/%s", repo.FullName, run.Name),
				Organization: s.config.Organization,
				Repository:   repo.Name,
				Data:         run,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("workflowRun", run.ID).WithError(err).Warn("failed to send workflow run import event")
			}
		}
	}

	return nil
}

func (s *GitHubSource) importIssues(ctx context.Context) error {
	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories for issues import: %w", err)
	}

	for _, repo := range repositories {
		issues, err := s.client.ListIssues(ctx, repo.Name)
		if err != nil {
			s.log.WithField("repository", repo.Name).WithError(err).Warn("failed to list issues for repository")
			continue
		}

		for _, issue := range issues {
			importEvent := GitHubImportEvent{
				Type:         "issue",
				ID:           issue.ID,
				Name:         fmt.Sprintf("Issue #%d", issue.Number),
				FullName:     fmt.Sprintf("%s#%d", repo.FullName, issue.Number),
				Organization: s.config.Organization,
				Repository:   repo.Name,
				Data:         issue,
			}

			if err := s.sendImportEvent(importEvent); err != nil {
				s.log.WithField("issue", issue.Number).WithError(err).Warn("failed to send issue import event")
			}
		}
	}

	return nil
}

func (s *GitHubSource) sendImportEvent(importEvent GitHubImportEvent) error {
	s.log.WithFields(logrus.Fields{
		"sourceType":   "github",
		"eventSource":  "import",
		"eventType":    importEvent.Type,
		"eventId":      importEvent.ID,
		"eventName":    importEvent.Name,
		"organization": importEvent.Organization,
	}).Debug("sending import event to pipeline")

	data, err := json.Marshal(importEvent)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import",
			"eventType":   importEvent.Type,
			"eventId":     importEvent.ID,
		}).WithError(err).Error("failed to marshal import event")
		return fmt.Errorf("failed to marshal import event: %w", err)
	}

	// Create a mock event builder similar to other sources
	eventBuilder := NewGitHubEventBuilder()
	event, err := eventBuilder.GetPipelineEvent(s.ctx, data)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"sourceType":  "github",
			"eventSource": "import",
			"eventType":   importEvent.Type,
			"eventId":     importEvent.ID,
		}).WithError(err).Error("failed to create pipeline event")
		return fmt.Errorf("failed to create pipeline event: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"sourceType":    "github",
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
	githubConfig, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return err
	}

	// Use the simple webhook setup for backward compatibility
	return webhook.SetupService(ctx, router, githubConfig.Configuration, pg)
}
