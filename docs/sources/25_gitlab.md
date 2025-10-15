````markdown
# GitLab

The GitLab source allows the integration-connector-agent to receive events from GitLab via webhooks and supports full import of GitLab resources.

## Webhook Integration

The GitLab source integrates with webhooks by exposing an endpoint at `/gitlab/webhook` (configurable).
When a webhook event is received, the following steps are performed:

1. **Validation**: The request is validated using the secret passed by the Webhook (token-based authentication, as per
  GitLab's requirements).
1. **Event Handling**: The event type is extracted from the `X-Gitlab-Event` header and injected into the event payload
  for routing. The event is then sent to the pipeline. The operation (e.g., `Write`) is determined based on the event
  type and action.

## Full Import

This source supports a full import of all GitLab resources in the configured group.
To trigger a full import, you can send a `POST` request to the import webhook path configured in the service configuration.

The full import includes:
- **Projects**: All projects in the group
- **Merge Requests**: All merge requests across all projects  
- **Pipelines**: All CI/CD pipelines across all projects
- **Releases**: All releases across all projects

### Service Configuration

The following configuration options are supported by the GitLab source:

- **type** (*string*): The type of the source, in this case `gitlab`
- **authentication** (*object*) *optional*: The authentication configuration for webhook events
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming webhook requests
- **webhookPath** (*string*) *optional*: The path where to receive the webhook events. Defaults to `/gitlab/webhook`.
- **token** ([*SecretSource*](../20_install.md#secretsource)) *optional*: GitLab personal access token for API access (required for import functionality)
- **baseUrl** (*string*) *optional*: GitLab instance base URL. Defaults to `https://gitlab.com`
- **group** (*string*) *optional*: GitLab group name (required for import functionality)
- **importWebhookPath** (*string*) *optional*: The path for the webhook exposed to trigger a full import
- **importAuthentication** (*object*) *optional*: The authentication configuration for import webhook
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming import webhook requests
  - **headerName** (*string*) *optional*: The name of the header used to validate incoming import webhook requests

#### Example - Basic Webhook Only

```json
{
  "type": "gitlab",
  "webhookPath": "/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITLAB_SECRET"
    }
  }
}
```

#### Example - With Full Import Support

```json
{
  "type": "gitlab",
  "webhookPath": "/gitlab/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITLAB_WEBHOOK_SECRET"
    }
  },
  "token": {
    "fromEnv": "GITLAB_API_TOKEN"
  },
  "baseUrl": "https://gitlab.com",
  "group": "my-group",
  "importWebhookPath": "/gitlab/import",
  "importAuthentication": {
    "secret": {
      "fromEnv": "GITLAB_IMPORT_SECRET"
    }
  }
}
```

#### Example - Self-hosted GitLab

```json
{
  "type": "gitlab",
  "webhookPath": "/gitlab/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "GITLAB_WEBHOOK_SECRET"
    }
  },
  "token": {
    "fromEnv": "GITLAB_API_TOKEN"
  },
  "baseUrl": "https://gitlab.example.com",
  "group": "my-organization",
  "importWebhookPath": "/gitlab/import",
  "importAuthentication": {
    "secret": {
      "fromEnv": "GITLAB_IMPORT_SECRET"
    }
  }
}
```

### How to Configure GitLab

To configure a webhook in GitLab, follow these steps:

1. **Group-level Webhook** (recommended for organization-wide events):
   - Navigate to your GitLab group
   - Go to Settings > Webhooks
   - Add webhook URL: `http://<your-agent-host>[/optional-base-path]/gitlab/webhook`
   - Set Secret Token (must match your configuration)
   - Select trigger events (see supported events below)

2. **Project-level Webhook** (for specific project events):
   - Navigate to your GitLab project
   - Go to Settings > Webhooks  
   - Add webhook URL: `http://<your-agent-host>[/optional-base-path]/gitlab/webhook`
   - Set Secret Token (must match your configuration)
   - Select trigger events (see supported events below)

Set the following fields:

- **URL**: The URL where the webhook will send events. For the GitLab integration, use `http://<your-agent-host>[/optional-base-path]/gitlab/webhook`.
- **Secret Token**: The secret used to validate incoming webhook requests. This must match the one set in the authentication configuration.
- **Trigger Events**: Select the events you want to subscribe to (see supported events section).
- **Enable SSL verification**: Enable if using HTTPS endpoints.

For full import functionality:

1. Create a GitLab Personal Access Token:
   - Go to User Settings > Access Tokens (or Group Settings > Access Tokens for group tokens)
   - Create token with appropriate scopes:
     - `read_api` scope for API access
     - `read_repository` scope for repository access
     - `read_user` scope for user information
2. Set the `token` in your configuration
3. Configure the group name in your configuration

## Supported Events

The GitLab source currently supports the following webhook events:

| Event                    | Event Type              | Example Payload                         | Operation |
|--------------------------|-------------------------|-----------------------------------------|-----------|
| Project Hook             | `Project Hook`          | [link](#project-event-payload)          | Write     |
| Merge Request Hook       | `Merge Request Hook`    | [link](#merge-request-event-payload)    | Write     |
| Pipeline Hook            | `Pipeline Hook`         | [link](#pipeline-event-payload)         | Write     |
| Release Hook             | `Release Hook`          | [link](#release-event-payload)          | Write     |
| Push Hook                | `Push Hook`             | [link](#push-event-payload)             | Write     |
| Tag Push Hook            | `Tag Push Hook`         | [link](#tag-push-event-payload)         | Write     |
| Issue Hook               | `Issue Hook`            | [link](#issue-event-payload)            | Write     |
| Note Hook                | `Note Hook`             | [link](#note-event-payload)             | Write     |
| Wiki Page Hook           | `Wiki Page Hook`        | [link](#wiki-page-event-payload)        | Write     |

The GitLab source supports the following resources for full import:

| Resource Type      | Import Type        | Description                                |
|--------------------|--------------------|--------------------------------------------|
| Project            | `project`          | All projects in the group                  |
| Merge Request      | `merge_request`    | All merge requests across projects         |
| Pipeline           | `pipeline`         | All CI/CD pipelines across projects       |
| Release            | `release`          | All releases across projects               |

:::info
The **event type** is extracted from the `X-Gitlab-Event` header and injected into the payload as `eventType` for
downstream processing.
:::

The operation is used by the sink to determine if the event should be inserted/updated or deleted.

### Example Payloads

#### Project Event Payload

The **event ID** used in the webhook payload is extracted from the `project.id` field.

The following is an example of a `Project Hook` event payload:

<details>
<summary>Project Event Payload</summary>

```json
{
  "object_kind": "project",
  "event_name": "project_create",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z",
  "project": {
    "id": 123,
    "name": "test-project",
    "description": "A test project for GitLab integration",
    "web_url": "https://gitlab.example.com/test-group/test-project",
    "path_with_namespace": "test-group/test-project",
    "default_branch": "main"
  },
  "user_id": 456,
  "user_name": "Test User",
  "user_username": "testuser"
}
```

</details>

#### Merge Request Event Payload

The **event ID** used in the webhook payload is extracted from the `object_attributes.id` field.

<details>
<summary>Merge Request Event Payload</summary>

```json
{
  "object_kind": "merge_request",
  "event_type": "merge_request",
  "user": {
    "id": 456,
    "name": "Test User",
    "username": "testuser"
  },
  "project": {
    "id": 123,
    "name": "test-project",
    "path_with_namespace": "test-group/test-project"
  },
  "object_attributes": {
    "id": 789,
    "iid": 1,
    "title": "Add new feature",
    "description": "This is a test merge request",
    "state": "opened",
    "merge_status": "can_be_merged",
    "target_branch": "main",
    "source_branch": "feature-branch",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

</details>

#### Import Event Payload

Import events have a standardized structure for all resource types:

<details>
<summary>Import Event Payload</summary>

```json
{
  "type": "project",
  "id": 123,
  "name": "test-project",
  "full_name": "test-group/test-project", 
  "group": "test-group",
  "data": {
    "id": 123,
    "name": "test-project",
    "path_with_namespace": "test-group/test-project",
    "description": "A test project for GitLab integration",
    "visibility": "private",
    "web_url": "https://gitlab.example.com/test-group/test-project",
    "default_branch": "main",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-12-01T00:00:00Z"
  }
}
```

</details>

### Extending Event Support

Refer to the [GitLab webhook events documentation](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#events)
for a full list of available events.
To add support to another event, open a pull request to [this repo](https://github.com/mia-platform/integration-connector-agent),
changing the [supported events mapping](https://github.com/mia-platform/integration-connector-agent/blob/main/internal/sources/gitlab/events.go).

````