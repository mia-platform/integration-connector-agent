# Repository

A GitHub Repository resource represents a code repository hosted on GitHub.
Repositories include metadata like name, description, owner, visibility, and timestamps.

## Mapping Example

```json
{
  "integrations": [
    {
      "source": {
        "type": "github",
        "webhookPath": "/github/webhook"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'repository'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "action": "{{ action }}",
                "repositoryId": "{{ repository.id }}",
                "repositoryName": "{{ repository.name }}",
                "repositoryFullName": "{{ repository.full_name }}",
                "repositoryDescription": "{{ repository.description }}",
                "repositoryLanguage": "{{ repository.language }}",
                "repositoryPrivate": "{{ repository.private }}",
                "repositoryHtmlUrl": "{{ repository.html_url }}",
                "repositoryCloneUrl": "{{ repository.clone_url }}",
                "repositoryDefaultBranch": "{{ repository.default_branch }}",
                "repositoryStarsCount": "{{ repository.stargazers_count }}",
                "repositoryForksCount": "{{ repository.forks_count }}",
                "repositoryOwnerLogin": "{{ repository.owner.login }}",
                "repositoryOwnerType": "{{ repository.owner.type }}",
                "repositoryCreatedAt": "{{ repository.created_at }}",
                "repositoryUpdatedAt": "{{ repository.updated_at }}",
                "senderLogin": "{{ sender.login }}",
                "senderType": "{{ sender.type }}",
                "organizationLogin": "{{ organization.login }}",
                "source": "github-webhook"
              }
            }
          ],
          "sinks": []
        }
      ]
    }
  ]
}

```
