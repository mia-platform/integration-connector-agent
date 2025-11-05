# GitHub

The GitHub source allows the integration-connector-agent to receive events from GitHub via webhooks and supports full import of GitHub resources.

## How to Configure GitHub

To configure a webhook in GitHub, follow the steps described in [GitHub's webhook documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks/creating-webhooks).

Set the following fields:

- **Payload URL**: The URL where the webhook will send events. For the GitHub integration, use `http://<your-agent-host>[/optional-base-path]/github/webhook`.
- **Content type**: `application/json` (recommended) or `application/x-www-form-urlencoded` (both are supported).
- **Secret**: The secret used to validate incoming webhook requests. This must match the one set in the authentication configuration.
- **Events**: Select the events you want to subscribe to.

For full import functionality, you can use either:

### Option 1: GitHub App Authentication (Recommended)

- Create a GitHub App in your organization:
  - Go to GitHub Settings > Developer settings > GitHub Apps
  - Click "New GitHub App"
  - Set the required permissions:
    - Repository permissions: Contents (Read), Metadata (Read), Pull requests (Read), Issues (Read), Actions (Read)
    - Organization permissions: Members (Read)
- Get the Client ID and Client Secret from your GitHub App
- Set the `clientId` and `clientSecret` in your configuration
- Configure the organization name in your configuration

### Option 2: Fine-Grained Access Token

- Create a GitHub Fine-Grained Access Token with appropriate access and permission:
  - Repository **access** `All repositories` scope for all future repositories owned. Also includes public repositories (read-only).
  - Repositories **permission** `Contents: Read-only` scope for repository contents, commits, branches, downloads, releases, and merges
  - Repositories **permission** `Metadata: Read-only` scope for search repositories, list collaborators, and access repository metadata
- Set the `token` in your configuration
- Configure the organization name in your configuration

## Supported Modes

### Webhook Events

The GitHub source currently supports the following webhook event that can be found in this [documentation](../processors/mappings/github/webhoook-event/5_supported_events.md).

### Full Import

The GitHub source supports the following resources for full import:

| Resource Type    | Import Type        | Description                          |
|------------------|--------------------|--------------------------------------|
| Repository       | `repository`       | All repositories in the organization |
| Pull Request     | `pull_request`     | All pull requests across repositories|
| Workflow Run     | `workflow_run`     | All GitHub Actions workflow runs     |
| Issue            | `issue`            | All issues across repositories       |

:::info
The **event type** is extracted from the `X-GitHub-Event` header and injected into the payload as `eventType` for
downstream processing.
:::

The operation is used by the sink to determine if the event should be inserted/updated or deleted.

## Service Configuration

The following configuration options are supported by the GitHub source:

- **type** (*string*): The type of the source, in this case `github`
- **authentication** (*object*) *optional*: The authentication configuration for webhook events
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming webhook requests
- **webhookPath** (*string*) *optional*: The path where to receive the webhook events. Defaults to `/github/webhook`.
- **clientId** ([*SecretSource*](../20_install.md#secretsource)) *optional*: GitHub App Client ID for API access (recommended for import functionality)
- **clientSecret** ([*SecretSource*](../20_install.md#secretsource)) *optional*: GitHub App Client Secret for API access (recommended for import functionality)
- **token** ([*SecretSource*](../20_install.md#secretsource)) *optional*: GitHub fine-grained access token for API access (better use clientId/clientSecret instead)
- **organization** (*string*) *optional*: GitHub organization name (required for import functionality)
- **importWebhookPath** (*string*) *optional*: The path for the webhook exposed to trigger a full import
- **importAuthentication** (*object*) *optional*: The authentication configuration for import webhook
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming import webhook requests
  - **headerName** (*string*) *optional*: The name of the header used to validate incoming import webhook requests

## Webhook Integration

The GitHub source integrates with webhooks by exposing an endpoint at `/github/webhook` (configurable).
When a webhook event is received, the following steps are performed:

- **Validation**: The request is validated using the secret passed by the Webhook (HMAC SHA256 signature, as per
  GitHub's requirements).
- **Event Handling**: The event type is extracted from the `X-GitHub-Event` header and injected into the event payload
  for routing. The event is then sent to the pipeline. The operation (e.g., `Write`) is determined based on the event
  type and action.

### Source Example - Basic Webhook Only

```json
{
  "type": "github",
  "webhookPath": "/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITHUB_SECRET"
    }
  }
}
```

## Full Import

This source supports a full import of all GitHub resources in the configured organization.
To trigger a full import, you can send a `POST` request to the import webhook path configured in the service configuration.

The full import includes:

- **Repositories**: All repositories in the organization
- **Pull Requests**: All pull requests across all repositories  
- **GitHub Actions**: All workflow runs across all repositories
- **Issues**: All issues across all repositories

### Source Example - With Full Import Support (GitHub App)

```json
{
  "type": "github",
  "webhookPath": "/github/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITHUB_WEBHOOK_SECRET"
    }
  },
  "clientId": {
    "fromEnv": "GITHUB_CLIENT_ID"
  },
  "clientSecret": {
    "fromEnv": "GITHUB_CLIENT_SECRET"
  },
  "organization": "my-organization",
  "importWebhookPath": "/github/import",
  "importAuthentication": {
    "secret": {
      "fromEnv": "GITHUB_IMPORT_SECRET"
    }
  }
}
```

### Source Example - With Full Import Support (Fine-Grained Token)

```json
{
  "type": "github",
  "webhookPath": "/github/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITHUB_WEBHOOK_SECRET"
    }
  },
  "token": {
    "fromEnv": "GITHUB_API_TOKEN"
  },
  "organization": "my-organization",
  "importWebhookPath": "/github/import",
  "importAuthentication": {
    "secret": {
      "fromEnv": "GITHUB_IMPORT_SECRET"
    }
  }
}
```

### Example Payloads

#### Webhook Event Payload

The **event ID** used in the webhook payload is extracted from the `pull_request.id` field.

The following is an example of a `pull_request` event payload:

```json
{
  "action": "opened",
  "number": 2,
  "pull_request": {
    "url": "https://api.github.com/repos/organization-name/project-name/pulls/2",
    "id": 2551578928,
    "html_url": "https://github.com/organization-name/project-name/pull/2",
    "title": "Create+test.json",
    "user": {
      "login": "johndoe",
      "id": 101523824
    },
    "body": "test+description",
    "created_at": "2025-05-29T08:53:54Z",
    ...
  },
  "repository": {
    "id": 983530734,
    "name": "project-name",
    "full_name": "organization-name/project-name"
  }
}
```

#### Import Event Payload

Import events have a standardized structure for all resource types:

```json
{
  "type": "repository",
  "id": 123456789,
  "name": "my-repository",
  "full_name": "organization-name/my-repository", 
  "organization": "organization-name",
  "data": {
    "id": 123456789,
    "name": "my-repository",
    "full_name": "organization-name/my-repository",
    "private": false,
    "html_url": "https://github.com/organization-name/my-repository",
    "language": "Go",
    "owner": {
      "login": "organization-name",
      "id": 987654321
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-12-01T00:00:00Z"
  }
}
```

### Example mappings

#### Webhook Events

There are some already prepared mappings for resources that can be used to import resources:

- [repository](../processors/mappings/github/webhoook-event/10_repository.md)
- [issue](../processors/mappings/github/webhoook-event/20_issue.md)
- [pull_request](../processors/mappings/github/webhoook-event/30_pull_request.md)
- [workflow_run](../processors/mappings/github/webhoook-event/40_workflow_run.md)
- [push](../processors/mappings/github/webhoook-event/50_push.md)
- [release](../processors/mappings/github/webhoook-event/60_release.md)

#### Full Import

There are some already prepared mappings for resources that can be used to import resources:

- [repository](../processors/mappings/github/rest-api/10_repository.md)
- [issue](../processors/mappings/github/rest-api/20_issue.md)
- [pull_request](../processors/mappings/github/rest-api/30_pull_request.md)
- [workflow_run](../processors/mappings/github/rest-api/40_workflow_run.md)

### Extending Event Support

Refer to the [GitHub webhook event types documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads)
for a full list of available events.
To add support to another event, open a pull request to [this repo](https://github.com/mia-platform/integration-connector-agent),
changing the [supported events mapping](https://github.com/mia-platform/integration-connector-agent/blob/main/internal/sources/github/events.go).
