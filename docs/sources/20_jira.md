# Jira

The Jira source supports to take events from Jira using the Jira Webhook.

## Webhook Integration

The Jira source integrates with webhooks by exposing an endpoint at `/jira/webhook`.
When a webhook event is received, the following steps are performed:

1. **Validation**: The request is validated using the secret passed by the Webhook.
1. **Event Handling**: The event type is extracted from the payload and the corresponding event is sent to the pipeline.
From the event type, it is also set which operation use: `Write` or `Delete` operation are supported by the sink.

### Service Configuration

The following configuration options are supported by the Jira source:

- **type** (*string*): The type of the source, in this case `jira`
- **authentication** (*object*) *optional*: The authentication configuration
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate the incoming webhook requests
- **webhookPath** (*string*) *optional*: The path where to receive the webhook events. Default to `/jira/webhook`.

#### Example

```json
{
  "type": "jira",
  "webhookPath": "/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "JIRA_SECRET"
    }
  }
}
```

### How to Configure Jira

To configure a webhook in Jira, follow the steps described in [this documentation](https://developer.atlassian.com/server/jira/platform/webhooks/).

As fields, you should set:

- **Name**: a name which identify the Webhook;
- **URL**: the URL where the Webhook will send the events. For the Jira integration, the URL should be `http://<your-agent-host>[/optional-base-path]/jira/webhook`;
- **Secret**: the secret used to validate the incoming webhook requests. This secret should be the same
as the one set in the authentication configuration.

## Supported Events

The Jira source supports the following webhook events:

- [issue events](#issue-events)

### Issue Events

For the issue events, it is possible to set a filter to receive only the events related to the issues that match the filter.

Example of a filter is `project = "My Project"`.

- issue created: `jira:issue_created`: this event will upsert data on the sink;
- issue updated: `jira:issue_updated`: this event will upsert data on the sink;
- issue deleted: `jira:issue_deleted`: this event will delete data on the sink.

:::info
The **event ID** used in the webhook payload is extracted from the `issue.id` field.
:::

#### Issue Event payload

The issue event payload is something like:

```json
{
  "id": 2,
  "timestamp": 1525698237764,
  "issue": {
    "id": "99291",
    "self": "https://jira.atlassian.com/rest/api/2/issue/99291",
    "key": "JRA-20002",
    "fields": {
      "summary": "I feel the need for speed",
      "created": "2009-12-16T23:46:10.612-0600",
      "description": "Make the issue nav load 10x faster",
      "labels": [
        "UI",
        "dialogue",
        "move"
      ],
      "priority": "Minor"
    }
  },
  "user": {
    "self": "https://jira.atlassian.com/rest/api/2/user?username=brollins",
    "name": "brollins",
    "key": "brollins",
    "emailAddress": "bryansemail at atlassian dot com",
    "avatarUrls": {
      "16x16": "https://jira.atlassian.com/secure/useravatar?size=small&avatarId=10605",
      "48x48": "https://jira.atlassian.com/secure/useravatar?avatarId=10605"
    },
    "displayName": "Bryan Rollins [Atlassian]",
    "active": "true"
  },
  "changelog": {
    "items": [
      {
        "toString": "A new summary.",
        "to": null,
        "fromString": "What is going on here?????",
        "from": null,
        "fieldtype": "jira",
        "field": "summary"
      },
      {
        "toString": "New Feature",
        "to": "2",
        "fromString": "Improvement",
        "from": "4",
        "fieldtype": "jira",
        "field": "issuetype"
      }
    ],
    "id": 10124
  },
  "comment": {
    "self": "https://jira.atlassian.com/rest/api/2/issue/10148/comment/252789",
    "id": "252789",
    "author": {
      "self": "https://jira.atlassian.com/rest/api/2/user?username=brollins",
      "name": "brollins",
      "emailAddress": "bryansemail@atlassian.com",
      "avatarUrls": {
        "16x16": "https://jira.atlassian.com/secure/useravatar?size=small&avatarId=10605",
        "48x48": "https://jira.atlassian.com/secure/useravatar?avatarId=10605"
      },
      "displayName": "Bryan Rollins [Atlassian]",
      "active": true
    },
    "body": "Just in time for AtlasCamp!",
    "updateAuthor": {
      "self": "https://jira.atlassian.com/rest/api/2/user?username=brollins",
      "name": "brollins",
      "emailAddress": "brollins@atlassian.com",
      "avatarUrls": {
        "16x16": "https://jira.atlassian.com/secure/useravatar?size=small&avatarId=10605",
        "48x48": "https://jira.atlassian.com/secure/useravatar?avatarId=10605"
      },
      "displayName": "Bryan Rollins [Atlassian]",
      "active": true
    },
    "created": "2011-06-07T10:31:26.805-0500",
    "updated": "2011-06-07T10:31:26.805-0500"
  },
  "webhookEvent": "jira:issue_updated"
}
```
