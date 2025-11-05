# Workflow Run

A GitHub Workflow Run represents a single execution of a workflow in GitHub Actions.
Workflow runs include status, triggering event, timestamps, and the user who triggered the run.

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
              "celExpression": "eventType == 'workflow_run'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "action": "{{ action }}",
                "workflowRunId": "{{ workflow_run.id }}",
                "workflowRunName": "{{ workflow_run.name }}",
                "workflowRunDisplayTitle": "{{ workflow_run.display_title }}",
                "workflowRunStatus": "{{ workflow_run.status }}",
                "workflowRunConclusion": "{{ workflow_run.conclusion }}",
                "workflowRunUrl": "{{ workflow_run.url }}",
                "workflowRunHtmlUrl": "{{ workflow_run.html_url }}",
                "workflowRunJobsUrl": "{{ workflow_run.jobs_url }}",
                "workflowRunLogsUrl": "{{ workflow_run.logs_url }}",
                "workflowRunCheckSuiteUrl": "{{ workflow_run.check_suite_url }}",
                "workflowRunArtifactsUrl": "{{ workflow_run.artifacts_url }}",
                "workflowRunCancelUrl": "{{ workflow_run.cancel_url }}",
                "workflowRunRerunUrl": "{{ workflow_run.rerun_url }}",
                "workflowRunWorkflowUrl": "{{ workflow_run.workflow_url }}",
                "workflowRunHeadBranch": "{{ workflow_run.head_branch }}",
                "workflowRunHeadSha": "{{ workflow_run.head_sha }}",
                "workflowRunRunNumber": "{{ workflow_run.run_number }}",
                "workflowRunRunAttempt": "{{ workflow_run.run_attempt }}",
                "workflowRunEvent": "{{ workflow_run.event }}",
                "workflowRunCreatedAt": "{{ workflow_run.created_at }}",
                "workflowRunUpdatedAt": "{{ workflow_run.updated_at }}",
                "workflowRunRunStartedAt": "{{ workflow_run.run_started_at }}",
                "workflowId": "{{ workflow.id }}",
                "workflowName": "{{ workflow.name }}",
                "workflowPath": "{{ workflow.path }}",
                "workflowState": "{{ workflow.state }}",
                "workflowUrl": "{{ workflow.url }}",
                "workflowHtmlUrl": "{{ workflow.html_url }}",
                "workflowBadgeUrl": "{{ workflow.badge_url }}",
                "workflowCreatedAt": "{{ workflow.created_at }}",
                "workflowUpdatedAt": "{{ workflow.updated_at }}",
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
