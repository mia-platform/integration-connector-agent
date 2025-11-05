# Workflow Run

A GitHub Workflow Run represents a single execution of a workflow in GitHub Actions.
Workflow runs include status, triggering event, timestamps, and the user who triggered the run.

The possible values a workflow run can have are:

- resourceType: the resource type ("workflow_run")
- id: unique numeric identifier for the workflow run
- name: human-readable name of the workflow
- fullName: repository full name with workflow name
- repository: repository name
- organization: organization or owner name
- status: run status (for example, completed)
- event: triggering event (for example, push)
- created_at: creation timestamp
- updated_at: last update timestamp
- user.id: actor user id
- user.login: actor username

## Mapping Example

```json
{
  "integrations": [
    {
      "source": {
        "type": "github",
        "importWebhookPath": "/github/import"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'github-import-workflow_run'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "resourceType": "{{ type }}",
                "id": "{{ id }}",
                "name": "{{ name }}",
                "fullName": "{{ full_name }}",
                "repository": "{{ repository }}",
                "organization": "{{ organization }}",
                "status": "{{ data.status }}",
                "event": "{{ data.event }}",
                "created_at": "{{ data.created_at }}",
                "updated_at": "{{ data.updated_at }}",
                "user": {
                  "id": "{{ data.actor.id }}",
                  "login": "{{ data.actor.login }}"
                }
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

## Data Example JSON representation

```json
{
  "resourceType": "workflow_run",
  "id": 18216294780,
  "name": "Node.js CI",
  "fullName": "mia-platform/custom-plugin-lib/Node.js CI",
  "repository": "custom-plugin-lib",
  "organization": "mia-platform",
  "status": "completed",
  "event": "push",
  "created_at": "2025-10-03T07:47:02Z",
  "updated_at": "2025-10-03T07:47:52Z",
  "user": {
    "id": 6539031,
    "login": "Pluto"
  }
}
```
