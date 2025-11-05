# Pull Request

A GitHub Pull Request represents a proposed code change submitted to a repository.
Pull requests include a head and base reference, metadata, and author information.

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
              "celExpression": "eventType == 'pull_request'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "action": "{{ action }}",
                "pullRequestId": "{{ pull_request.id }}",
                "pullRequestNumber": "{{ pull_request.number }}",
                "pullRequestTitle": "{{ pull_request.title }}",
                "pullRequestBody": "{{ pull_request.body }}",
                "pullRequestState": "{{ pull_request.state }}",
                "pullRequestDraft": "{{ pull_request.draft }}",
                "pullRequestMerged": "{{ pull_request.merged }}",
                "pullRequestMergeable": "{{ pull_request.mergeable }}",
                "pullRequestRebaseable": "{{ pull_request.rebaseable }}",
                "pullRequestHtmlUrl": "{{ pull_request.html_url }}",
                "pullRequestUrl": "{{ pull_request.url }}",
                "pullRequestDiffUrl": "{{ pull_request.diff_url }}",
                "pullRequestPatchUrl": "{{ pull_request.patch_url }}",
                "pullRequestCommits": "{{ pull_request.commits }}",
                "pullRequestAdditions": "{{ pull_request.additions }}",
                "pullRequestDeletions": "{{ pull_request.deletions }}",
                "pullRequestChangedFiles": "{{ pull_request.changed_files }}",
                "pullRequestCreatedAt": "{{ pull_request.created_at }}",
                "pullRequestUpdatedAt": "{{ pull_request.updated_at }}",
                "pullRequestClosedAt": "{{ pull_request.closed_at }}",
                "pullRequestMergedAt": "{{ pull_request.merged_at }}",
                "pullRequestAuthorLogin": "{{ pull_request.user.login }}",
                "pullRequestAuthorId": "{{ pull_request.user.id }}",
                "pullRequestAuthorType": "{{ pull_request.user.type }}",
                "pullRequestAuthorAssociation": "{{ pull_request.author_association }}",
                "pullRequestHeadRef": "{{ pull_request.head.ref }}",
                "pullRequestHeadSha": "{{ pull_request.head.sha }}",
                "pullRequestHeadLabel": "{{ pull_request.head.label }}",
                "pullRequestBaseRef": "{{ pull_request.base.ref }}",
                "pullRequestBaseSha": "{{ pull_request.base.sha }}",
                "pullRequestBaseLabel": "{{ pull_request.base.label }}",
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
