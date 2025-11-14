// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package github

import (
	"github.com/gofiber/fiber/v2"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/tidwall/gjson"
)

const (
	githubEventHeader = "X-GitHub-Event"

	// Repository events
	repositoryEvent = "repository"

	// Pull request events
	pullRequestEvent = "pull_request"

	// Issue events
	issuesEvent = "issues"

	// Release events
	releaseEvent = "release"

	// Workflow events
	workflowRunEvent = "workflow_run"
	workflowJobEvent = "workflow_job"

	// Deployment events
	deploymentEvent = "deployment"

	// Label events
	labelEvent = "label"

	// Package events
	packageEvent = "package"

	// Personal access token request events
	personalAccessTokenRequestEvent = "personal_access_token_request"

	// Security and analysis events
	repositoryAdvisoryEvent = "repository_advisory"
)

func getPrimaryKeyFromPathsArray(pathsArray []string) func(parsedData gjson.Result) entities.PkFields {
	return func(parsedData gjson.Result) entities.PkFields {
		if len(pathsArray) < 1 {
			return nil
		}

		pkFieldsArray := make(entities.PkFields, 0, len(pathsArray))
		for _, path := range pathsArray {
			if parsedData.Get(path).String() == "" {
				return nil
			}
			pkFieldsArray = append(pkFieldsArray, entities.PkField{Key: path, Value: parsedData.Get(path).String()})
		}

		return pkFieldsArray
	}
}

var SupportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		// Repository events
		repositoryEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Pull request events
		pullRequestEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("pull_request.id"),
		},

		// Issue events
		issuesEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"issue.id", "repository.id"}),
		},

		// Release events
		releaseEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"release.id", "repository.id"}),
		},

		// Workflow events
		workflowRunEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"workflow_run.id", "workflow.id"}),
		},
		workflowJobEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"workflow_job.id", "workflow.id"}),
		},

		// Deployment events
		deploymentEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"deployment.id", "repository.id"}),
		},

		// Label events
		labelEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"label.id", "repository.id"}),
		},

		// Package events
		packageEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"package.id", "package.namespace"}),
		},

		// Personal access token request events
		personalAccessTokenRequestEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"personal_access_token_request.id", "personal_access_token_request.token_id"}),
		},

		// Repository advisory events
		repositoryAdvisoryEvent: {
			Operation:  entities.Write,
			GetFieldID: getPrimaryKeyFromPathsArray([]string{"repository_advisory.ghsa_id", "repository.id"}),
		},
	},
	GetEventType: func(data webhook.EventTypeParam) string {
		return data.Headers.Get(githubEventHeader)
	},
	PayloadKey: webhook.ContentTypeConfig{
		fiber.MIMEApplicationForm: "payload",
	},
}
