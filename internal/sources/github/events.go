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

package github

import (
	"github.com/gofiber/fiber/v2"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	githubEventHeader = "X-GitHub-Event"

	// Repository events
	repositoryEvent = "repository"

	// Pull request events
	pullRequestEvent       = "pull_request"
	pullRequestReviewEvent = "pull_request_review"

	// Issue events
	issuesEvent       = "issues"
	issueCommentEvent = "issue_comment"

	// Push and commit events
	pushEvent   = "push"
	createEvent = "create"
	deleteEvent = "delete"

	// Release events
	releaseEvent = "release"

	// Fork and star events
	forkEvent  = "fork"
	watchEvent = "watch"

	// Branch and tag events
	branchProtectionRuleEvent = "branch_protection_rule"

	// Workflow events
	workflowRunEvent = "workflow_run"
	workflowJobEvent = "workflow_job"

	// Security events
	secretScanningAlertEvent = "secret_scanning_alert"
	codeScanningAlertEvent   = "code_scanning_alert"
	dependabotAlertEvent     = "dependabot_alert"

	// Collaboration events
	memberEvent     = "member"
	membershipEvent = "membership"
	teamEvent       = "team"
	teamAddEvent    = "team_add"

	// Project events
	projectEvent       = "project"
	projectCardEvent   = "project_card"
	projectColumnEvent = "project_column"

	// Wiki and pages events
	gollumEvent     = "gollum"
	pagesBuildEvent = "page_build"

	// Deployment events
	deploymentEvent       = "deployment"
	deploymentStatusEvent = "deployment_status"

	// Status events
	statusEvent     = "status"
	checkRunEvent   = "check_run"
	checkSuiteEvent = "check_suite"

	// Discussion events
	discussionEvent        = "discussion"
	discussionCommentEvent = "discussion_comment"

	// Milestone events
	milestoneEvent = "milestone"

	// Label events
	labelEvent = "label"

	// Organization events
	organizationEvent = "organization"

	// GitHub App events
	installationEvent             = "installation"
	installationRepositoriesEvent = "installation_repositories"
	githubAppAuthorizationEvent   = "github_app_authorization"

	// Marketplace events
	marketplacePurchaseEvent = "marketplace_purchase"

	// Sponsorship events
	sponsorshipEvent = "sponsorship"

	// Package events
	packageEvent = "package"

	// Registry package events
	registryPackageEvent = "registry_package"

	// Meta events
	metaEvent = "meta"

	// Ping event (webhook test)
	pingEvent = "ping"

	// Star events (GitHub's new star event)
	starEvent = "star"

	// Repository vulnerability alert events
	repositoryVulnerabilityAlertEvent = "repository_vulnerability_alert"

	// Repository dispatch events
	repositoryDispatchEvent = "repository_dispatch"

	// Workflow dispatch events
	workflowDispatchEvent = "workflow_dispatch"

	// Personal access token request events
	personalAccessTokenRequestEvent = "personal_access_token_request"

	// Projects v2 events (GitHub's new project boards)
	projectsV2Event     = "projects_v2"
	projectsV2ItemEvent = "projects_v2_item"

	// Security and analysis events
	repositoryAdvisoryEvent = "repository_advisory"

	// Code security and analysis
	codeOwnershipEvent = "code_ownership"

	// Enterprise events
	enterpriseEvent = "enterprise"

	// Global security events
	globalSecurityEvent = "global_security"
)

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
		pullRequestReviewEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("review.id"),
		},

		// Issue events
		issuesEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("issue.id"),
		},
		issueCommentEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("comment.id"),
		},

		// Push and commit events
		pushEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},
		createEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},
		deleteEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Release events
		releaseEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("release.id"),
		},

		// Fork and star events
		forkEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("forkee.id"),
		},
		watchEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Branch protection events
		branchProtectionRuleEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("rule.id"),
		},

		// Workflow events
		workflowRunEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("workflow_run.id"),
		},
		workflowJobEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("workflow_job.id"),
		},

		// Security events
		secretScanningAlertEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("alert.number"),
		},
		codeScanningAlertEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("alert.number"),
		},
		dependabotAlertEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("alert.number"),
		},

		// Collaboration events
		memberEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("member.id"),
		},
		membershipEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("member.id"),
		},
		teamEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("team.id"),
		},
		teamAddEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("team.id"),
		},

		// Project events
		projectEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project.id"),
		},
		projectCardEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project_card.id"),
		},
		projectColumnEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("project_column.id"),
		},

		// Wiki and pages events
		gollumEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},
		pagesBuildEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("id"),
		},

		// Deployment events
		deploymentEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("deployment.id"),
		},
		deploymentStatusEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("deployment_status.id"),
		},

		// Status events
		statusEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("id"),
		},
		checkRunEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("check_run.id"),
		},
		checkSuiteEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("check_suite.id"),
		},

		// Discussion events
		discussionEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("discussion.id"),
		},
		discussionCommentEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("comment.id"),
		},

		// Milestone events
		milestoneEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("milestone.id"),
		},

		// Label events
		labelEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("label.id"),
		},

		// Organization events
		organizationEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("organization.id"),
		},

		// GitHub App events
		installationEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("installation.id"),
		},
		installationRepositoriesEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("installation.id"),
		},
		githubAppAuthorizationEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("sender.id"),
		},

		// Marketplace events
		marketplacePurchaseEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("marketplace_purchase.id"),
		},

		// Sponsorship events
		sponsorshipEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("sponsorship.id"),
		},

		// Package events
		packageEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("package.id"),
		},

		// Registry package events
		registryPackageEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("registry_package.id"),
		},

		// Meta events
		metaEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("hook.id"),
		},

		// Ping event (web hook test)
		pingEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("hook.id"),
		},

		// Star events
		starEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Repository vulnerability alert events
		repositoryVulnerabilityAlertEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("alert.id"),
		},

		// Repository dispatch events
		repositoryDispatchEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Workflow dispatch events
		workflowDispatchEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Personal access token request events
		personalAccessTokenRequestEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("personal_access_token_request.id"),
		},

		// Projects v2 events
		projectsV2Event: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("projects_v2.id"),
		},
		projectsV2ItemEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("projects_v2_item.id"),
		},

		// Repository advisory events
		repositoryAdvisoryEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository_advisory.id"),
		},

		// Code ownership events
		codeOwnershipEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},

		// Enterprise events
		enterpriseEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("enterprise.id"),
		},

		// Global security events
		globalSecurityEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("repository.id"),
		},
	},
	GetEventType: func(data webhook.EventTypeParam) string {
		return data.Headers.Get(githubEventHeader)
	},
	PayloadKey: webhook.ContentTypeConfig{
		fiber.MIMEApplicationForm: "payload",
	},
}
