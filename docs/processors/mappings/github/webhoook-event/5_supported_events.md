# Supported GitHub webhook events

This document lists the GitHub webhook events supported by the GitHub source. For each event we show:

- Event name (the value of the `X-GitHub-Event` header)
- A short description of the event payload / when it's sent
- The operation used by the pipeline (Write or Delete)
- The JSON path used to extract the primary ID from the payload (used as the event ID)

Note: The agent injects the event type (from the `X-GitHub-Event` header)
into the payload as `eventType` for downstream processing.

Payloads sent with content-type `application/x-www-form-urlencoded` are
expected to have the JSON body in the `payload` form key.

## Supported events

| Event name | Description | Operation | Primary keys |
|---|---|---:|---|
| repository | Sent when a repository is created, deleted, archived, renamed, etc. | Write | `repository.id` |
| pull_request | Sent for pull request lifecycle events (opened, closed, synchronized, etc.) | Write | `pull_request.id` |
| issues | Sent when an issue is created, edited, closed, reopened, etc. | Write | `issue.id` and `repository.id` |
| release | Sent when a release is published, edited, or deleted | Write | `release.id` and `repository.id` |
| workflow_run | Sent for GitHub Actions workflow run lifecycle events | Write | `workflow_run.id` and `workflow.id` |
| workflow_job | Sent for individual workflow job lifecycle events | Write | `workflow_job.id` and `workflow.id` |
| deployment | Sent when a deployment is created | Write | `deployment.id` and `repository.id` |
| label | Label created/updated events | Write | `label.id` and `repository.id` |
| package | Package publish/update events | Write | `package.id` |
| personal_access_token_request | Events about personal access token requests | Write | `personal_access_token_request.id` and `personal_access_token_request.token_id` |
| repository_advisory | Repository advisory events | Write | `repository_advisory.ghsa_id` and `repository.id` |

## Notes

- If the event payload does not contain the expected primary key at the
  specified path, the event ID extraction will fail. The event may be
  skipped or produce an error depending on pipeline configuration.
- For webhook validation the agent expects the webhook secret to be
  configured (HMAC SHA256 signature verification). See the GitHub source
  documentation for authentication and webhook setup.
- To add or change supported events, update
  `internal/sources/github/events.go` and open a pull request.
