# GitLab Mia-Platform Item Types

This directory contains JSON schema definitions for GitLab resources that can be integrated with Mia-Platform Console using the integration-connector-agent.

## Item Types

### gitlab-project.json
Schema for GitLab projects/repositories. Includes project metadata, settings, namespace information, and repository details.

**Key Fields:**
- `projectId`: Unique project identifier
- `projectName`: Project name
- `projectFullName`: Full project path (namespace/project)
- `projectDescription`: Project description
- `projectVisibility`: Project visibility level (private/internal/public)
- `projectWebUrl`: Project web interface URL
- `projectHttpUrlToRepo`: HTTP clone URL
- `projectSshUrlToRepo`: SSH clone URL

### gitlab-merge-request.json
Schema for GitLab merge requests. Includes merge request metadata, state, author information, and project references.

**Key Fields:**
- `mergeRequestId`: Unique merge request identifier
- `mergeRequestIid`: Internal merge request ID (per project)
- `mergeRequestTitle`: Merge request title
- `mergeRequestState`: Current state (opened/closed/merged/locked)
- `mergeRequestMergeStatus`: Merge status (can_be_merged/cannot_be_merged/etc.)
- `mergeRequestTargetBranch`: Target branch name
- `mergeRequestSourceBranch`: Source branch name

### gitlab-pipeline.json
Schema for GitLab CI/CD pipelines. Includes pipeline execution details, status, and associated project information.

**Key Fields:**
- `pipelineId`: Unique pipeline identifier
- `pipelineRef`: Git reference (branch/tag)
- `pipelineSha`: Commit SHA
- `pipelineStatus`: Current status (running/success/failed/etc.)
- `pipelineSource`: Pipeline trigger source (push/web/trigger/etc.)
- `pipelineDuration`: Execution duration in seconds

### gitlab-release.json
Schema for GitLab releases. Includes release metadata, associated commit information, and asset details.

**Key Fields:**
- `releaseTagName`: Git tag name
- `releaseName`: Release name/title
- `releaseDescription`: Release description/notes
- `releaseCreatedAt`: Creation timestamp
- `releaseReleasedAt`: Publication timestamp
- `releaseCommitId`: Associated commit SHA

## Usage

These schemas define the structure of events that will be processed by the integration-connector-agent when receiving GitLab webhook events or performing GitLab imports. Each schema corresponds to a specific GitLab resource type and includes all relevant fields for integration with Mia-Platform Console.

## Source Types

All schemas support two source types:
- `gitlab-webhook`: Events received via GitLab webhooks
- `gitlab-import`: Events generated during bulk import operations

## Common Fields

All schemas include these common fields:
- `eventType`: Type of GitLab event
- `action`: Action performed (create/update/delete/import/etc.)
- `projectId`: Associated project ID
- `group`: GitLab group/namespace
- `userLogin`: Event triggering user
- `source`: Data source type