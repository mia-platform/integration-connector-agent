# Microsoft Azure DevOps

The Microsoft Azure DevOps source allows the integration-connector-agent to receive events from Azure DevOps.

## Full Import

This source supports a full import of all the git repositories reachable with the configured Azure DevOps PAT.  
To trigger a full import, you can send a `POST` request to the import webhook path configured in the service
configuration.

## Webhook Integration

The Jira source integrates with webhooks by exposing an endpoint at `/azure-devops/webhook`. When a webhook event is
received, the following steps are performed:

1. **Validation**: The request is validated using the secret passed by the Webhook.
1. **Event Handling**: The event type is extracted from the payload and the corresponding event is sent to the pipeline.
From the event type, it is also set which operation use: `Write` or `Delete` operation are supported by the sink.

### Service Configuration

The following configuration options are supported by the GitHub source:

- **type** (*string*): The type of the source, in this case `azure-devops`
- **authentication** (*object*) *optional*: The authentication configuration
  - **username** (*string*): The username used to validate incoming webhook requests
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming webhook requests
- **webhookPath** (*string*) *optional*: The path where to receive the webhook events. Defaults to `/azure-devops/webhook`.
- **organizationSubscriptions** (*boolean*) *optional*: if set to true the Azure DevOps PAT must have admin privileges
	and the webhooks subscriptions will be registered on the organization itself receiving events from all the projects
- **azureDevOpsOrganizationUrl** (*string*): The Azure DevOps organization URL
- **azureDevOpsPersonalAccessToken** ([*SecretSource*](../20_install.md#secretsource)): The PAT used to authorize the
  calls to th Azure DevOps endpoint
- **importWebhookPath** (*string*) *optional*: The path to enable the webhook to trigger a full load from Azure DevOps

#### Example

```json
{
  "type": "azure-devops",
  "webhookPath": "/webhook",
  "webhookHost": "http://example.com",
  "authentication": {
    "username": "user",
    "secret": {
      "fromEnv": "SECRET"
    }
  },
  "azureDevOpsOrganizationUrl": "https://http://dev.azure.com/organizationName",
  "azureDevOpsPersonalAccessToken": {
    "fromEnv": "PAT_SECRET"
  }
}
```

### How to Configure Azure DevOps

The sink will auto configure the relevant service hooks during startup on all the projects that can be read via the
Azure DevOps PAT used in the configuration.

## Supported Events

The Azure DevOps source supports the following webhook events:

| Event              | Event Type         | Example Payload                                   | Operation |
|--------------------|--------------------|---------------------------------------------------|-----------|
| repository created | `git.repo.created` | [Git Repository Created](#git-repository-created) | Write     |
| repository renamed | `git.repo.renamed` | [Git Repository Renamed](#git-repository-renamed) | Write     |
| repository deleted | `git.repo.deleted` | [Git Repository Deleted](#git-repository-deleted) | Delete    |

The operation will be used by the sink which supports the upsert of the data to decide if
the event should be inserted/updated or deleted.

### Example Payloads

#### Git Repository Created

The **repository ID** used in the webhook payload is extracted from the `resource.repository.id` field.

<details>
<summary>Git Repository Created Event Payload</summary>

```json
{
  "id": "a0a0a0a0-bbbb-cccc-dddd-e1e1e1e1e1e1",
  "eventType": "git.repo.created",
  "publisherId": "tfs",
  "message": {
    "text": "A new Git repository was created with name Fabrikam-Fiber-Git and ID c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3.",
    "html": "A new Git repository was created with name <a href=\"https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git/\">Fabrikam-Fiber-Git</a> and ID c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3.",
    "markdown": "A new Git repository was created with name [Fabrikam-Fiber-Git](https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git/) and ID `c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3`."
  },
  "detailedMessage": {
    "text": "A new Git repository was created with name Fabrikam-Fiber-Git and ID c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3.",
    "html": "A new Git repository was created with name <a href=\"https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git/\">Fabrikam-Fiber-Git</a> and ID c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3.",
    "markdown": "A new Git repository was created with name [Fabrikam-Fiber-Git](https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git/) and ID `c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3`."
  },
  "resource": {
    "repository": {
      "id": "c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3",
      "name": "Fabrikam-Fiber-Git",
      "url": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3",
      "project": {
        "id": "00aa00aa-bb11-cc22-dd33-44ee44ee44ee",
        "name": "Fabrikam-Fiber-Git",
        "url": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/projects/00aa00aa-bb11-cc22-dd33-44ee44ee44ee",
        "state": "wellFormed",
        "revision": 11,
        "visibility": "private",
        "lastUpdateTime": "2025-06-12T20:22:53.7494088+00:00"
      },
      "defaultBranch": "refs/heads/main",
      "size": 728,
      "remoteUrl": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git",
      "sshUrl": "ssh://git@ssh.fabrikam-fiber-inc.visualstudio.com/v3/DefaultCollection/Fabrikam-Fiber-Git",
      "isDisabled": false
    },
    "initiatedBy": {
      "displayName": "Ivan Yurev",
      "id": "22cc22cc-dd33-ee44-ff55-66aa66aa66aa",
      "uniqueName": "user@fabrikamfiber.com"
    },
    "utcTimestamp": "2022-12-12T12:34:56.5498459Z"
  },
  "resourceVersion": "1.0-preview.1",
  "resourceContainers": {
    "collection": {
      "id": "b1b1b1b1-cccc-dddd-eeee-f2f2f2f2f2f2"
    },
    "account": {
      "id": "bbbb1b1b-cc2c-dd3d-ee4e-ffffff5f5f5f"
    },
    "project": {
      "id": "00aa00aa-bb11-cc22-dd33-44ee44ee44ee"
    }
  },
  "createdDate": "2025-06-12T20:22:53.818Z"
}
```

</details>

#### Git Repository Renamed

The **repository ID** used in the webhook payload is extracted from the `resource.repository.id` field.

<details>
<summary>Git Repository Renamed Event Payload</summary>

```json
{
  "id": "a0a0a0a0-bbbb-cccc-dddd-e1e1e1e1e1e1",
  "eventType": "git.repo.renamed",
  "publisherId": "tfs",
  "message": {
    "text": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to Fabrikam-Fiber-Git.",
    "html": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to  <a href=\"https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3\">Fabrikam-Fiber-Git</a>.",
    "markdown": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to [Fabrikam-Fiber-Git](https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3)."
  },
  "detailedMessage": {
    "text": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to Fabrikam-Fiber-Git.\r\nProject name: Contoso\r\n\r\nRepository name before renaming: Diber-Git\r\n\r\nDefault branch: refs/heads/main\r\n\r\nRepository link(https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3)\r\n",
    "html": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to  <a href=\"https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3\">Fabrikam-Fiber-Git</a>.<p>Project name: Contoso</p><p>Repository name before renaming: Diber-Git</p><p>Default branch: refs/heads/main</p><p><a href=\"https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3\">Repository link</a></p>",
    "markdown": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was renamed to [Fabrikam-Fiber-Git](https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3).\r\nProject name: Contoso\r\n\r\nRepository name before renaming: Diber-Git\r\n\r\nDefault branch: refs/heads/main\r\n\r\n[Repository link](https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3)\r\n"
  },
  "resource": {
    "oldName": "Diber-Git",
    "newName": "Fabrikam-Fiber-Git",
    "repository": {
      "id": "c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3",
      "name": "Fabrikam-Fiber-Git",
      "url": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/git/repositories/c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3",
      "project": {
        "id": "11bb11bb-cc22-dd33-ee44-55ff55ff55ff",
        "name": "Contoso",
        "url": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/projects/11bb11bb-cc22-dd33-ee44-55ff55ff55ff",
        "state": "wellFormed",
        "revision": 11,
        "visibility": "private",
        "lastUpdateTime": "2025-06-12T20:48:38.8174565+00:00"
      },
      "defaultBranch": "refs/heads/main",
      "size": 728,
      "remoteUrl": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_git/Fabrikam-Fiber-Git",
      "sshUrl": "ssh://git@ssh.fabrikam-fiber-inc.visualstudio.com/v3/DefaultCollection/Fabrikam-Fiber-Git",
      "isDisabled": false
    },
    "initiatedBy": {
        "displayName": "Himani Maharjan",
        "id": "a0a0a0a0-bbbb-cccc-dddd-e1e1e1e1e1e1",
        "uniqueName": "himani@fabrikamfiber.com"
    },
    "utcTimestamp": "2022-12-12T12:34:56.5498459Z"
  },
  "resourceVersion": "1.0-preview.1",
  "resourceContainers": {
    "collection": {
      "id": "b1b1b1b1-cccc-dddd-eeee-f2f2f2f2f2f2"
    },
    "account": {
      "id": "bbbb1b1b-cc2c-dd3d-ee4e-ffffff5f5f5f"
    },
    "project": {
      "id": "00aa00aa-bb11-cc22-dd33-44ee44ee44ee"
    }
  },
  "createdDate": "2025-06-12T20:48:38.859Z"
}
```

</details>

#### Git Repository Deleted

The **repository ID** used in the webhook payload is extracted from the `resource.repositoryId` field.

<details>
<summary>Git Repository Deleted Event Payload</summary>

```json
{
  "id": "a0a0a0a0-bbbb-cccc-dddd-e1e1e1e1e1e1",
  "eventType": "git.repo.deleted",
  "publisherId": "tfs",
  "message": {
    "text": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted.",
    "html": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted.",
    "markdown": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted."
  },
  "detailedMessage": {
    "text": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted.\r\nProject name: Contoso\r\n\r\nRepository name: Fabrikam-Fiber-Git\r\n\r\nRepository can be restored: true\r\n",
    "html": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted.<p>Project name: Contoso</p><p>Repository name: Fabrikam-Fiber-Git</p><p>Repository can be restored: true</p>",
    "markdown": "Git repository c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3 was deleted.\r\nProject name: Contoso\r\n\r\nRepository name: Fabrikam-Fiber-Git\r\n\r\nRepository can be restored: true\r\n"
  },
  "resource": {
    "project": {
      "id": "00aa00aa-bb11-cc22-dd33-44ee44ee44ee",
      "name": "Contoso",
      "url": "https://fabrikam-fiber-inc.visualstudio.com/DefaultCollection/_apis/projects/00aa00aa-bb11-cc22-dd33-44ee44ee44ee",
      "state": "wellFormed",
      "revision": 11,
      "visibility": "private",
      "lastUpdateTime": "2025-06-12T20:33:32.4370396+00:00"
    },
    "repositoryId": "c2c2c2c2-dddd-eeee-ffff-a3a3a3a3a3a3",
    "repositoryName": "Fabrikam-Fiber-Git",
    "isHardDelete": false,
    "initiatedBy": {
      "displayName": "Himani Maharjan",
      "id": "d3d3d3d3-eeee-ffff-aaaa-b4b4b4b4b4b4",
      "uniqueName": "himani@fabrikamfiber.com"
    },
    "utcTimestamp": "2022-12-12T12:34:56.5498459Z"
  },
  "resourceVersion": "1.0-preview.1",
  "resourceContainers": {
    "collection": {
      "id": "b1b1b1b1-cccc-dddd-eeee-f2f2f2f2f2f2"
    },
    "account": {
      "id": "bbbb1b1b-cc2c-dd3d-ee4e-ffffff5f5f5f"
    },
    "project": {
      "id": "00aa00aa-bb11-cc22-dd33-44ee44ee44ee"
    }
  },
  "createdDate": "2025-06-12T20:33:32.512Z"
}
```

</details>
