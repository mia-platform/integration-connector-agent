# Push

A GitHub Push event on a repository hosted on GitHub.

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
              "celExpression": "eventType == 'push'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "ref": "{{ ref }}",
                "refType": "{{ ref_type }}",
                "before": "{{ before }}",
                "after": "{{ after }}",
                "created": "{{ created }}",
                "deleted": "{{ deleted }}",
                "forced": "{{ forced }}",
                "baseRef": "{{ base_ref }}",
                "compare": "{{ compare }}",
                "commits": "{{ commits }}",
                "headCommitId": "{{ head_commit.id }}",
                "headCommitTreeId": "{{ head_commit.tree_id }}",
                "headCommitMessage": "{{ head_commit.message }}",
                "headCommitTimestamp": "{{ head_commit.timestamp }}",
                "headCommitUrl": "{{ head_commit.url }}",
                "headCommitAuthorName": "{{ head_commit.author.name }}",
                "headCommitAuthorEmail": "{{ head_commit.author.email }}",
                "headCommitAuthorUsername": "{{ head_commit.author.username }}",
                "headCommitCommitterName": "{{ head_commit.committer.name }}",
                "headCommitCommitterEmail": "{{ head_commit.committer.email }}",
                "headCommitCommitterUsername": "{{ head_commit.committer.username }}",
                "pusherName": "{{ pusher.name }}",
                "pusherEmail": "{{ pusher.email }}",
                "repositoryId": "{{ repository.id }}",
                "repositoryName": "{{ repository.name }}",
                "repositoryFullName": "{{ repository.full_name }}",
                "repositoryPrivate": "{{ repository.private }}",
                "repositoryHtmlUrl": "{{ repository.html_url }}",
                "repositoryOwnerLogin": "{{ repository.owner.login }}",
                "repositoryOwnerType": "{{ repository.owner.type }}",
                "organizationLogin": "{{ organization.login }}",
                "senderLogin": "{{ sender.login }}",
                "senderType": "{{ sender.type }}",
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
