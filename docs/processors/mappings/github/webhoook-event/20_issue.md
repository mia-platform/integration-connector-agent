# Issue

A GitHub Issue represents a discussion thread in a repository.
Issues are used to track bugs, feature requests, and other work items.

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
              "celExpression": "eventType == 'issues'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "action": "{{ action }}",
                "issueId": "{{ issue.id }}",
                "issueNumber": "{{ issue.number }}",
                "issueTitle": "{{ issue.title }}",
                "issueBody": "{{ issue.body }}",
                "issueState": "{{ issue.state }}",
                "issueStateReason": "{{ issue.state_reason }}",
                "issueLocked": "{{ issue.locked }}",
                "issueHtmlUrl": "{{ issue.html_url }}",
                "issueUrl": "{{ issue.url }}",
                "issueCommentsUrl": "{{ issue.comments_url }}",
                "issueEventsUrl": "{{ issue.events_url }}",
                "issueLabelsUrl": "{{ issue.labels_url }}",
                "issueTimelineUrl": "{{ issue.timeline_url }}",
                "issueCreatedAt": "{{ issue.created_at }}",
                "issueUpdatedAt": "{{ issue.updated_at }}",
                "issueClosedAt": "{{ issue.closed_at }}",
                "issueAuthorLogin": "{{ issue.user.login }}",
                "issueAuthorId": "{{ issue.user.id }}",
                "issueAuthorType": "{{ issue.user.type }}",
                "issueAuthorAssociation": "{{ issue.author_association }}",
                "issueAssigneeLogin": "{{ issue.assignee.login }}",
                "issueAssigneeId": "{{ issue.assignee.id }}",
                "issueLabels": "{{ issue.labels }}",
                "issueMilestoneTitle": "{{ issue.milestone.title }}",
                "issueMilestoneId": "{{ issue.milestone.id }}",
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
