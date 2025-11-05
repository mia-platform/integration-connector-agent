# Issue

A GitHub Issue represents a discussion thread in a repository.
Issues are used to track bugs, feature requests, and other work items.

The possible values an issue can have are:

- resourceType: the resource type (for example, "issue")
- id: unique numeric identifier for the issue
- name: human-readable name
- fullName: repository full name with issue number
- repository: repository name
- organization: organization or owner name
- title: issue title
- number: issue number within the repository
- state: issue state (open, closed)
- labels: list of labels attached to the issue
- created_at: creation timestamp
- updated_at: last update timestamp
- user.id: author user id
- user.login: author username

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
              "celExpression": "eventType == 'github-import-issue'"
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
                "title": "{{ data.title }}",
                "number": "{{ data.number }}",
                "state": "{{ data.state }}",
                "labels": "{{ data.labels }}",
                "created_at": "{{ data.created_at }}",
                "updated_at": "{{ data.updated_at }}",
                "user": {
                  "id": "{{ data.user.id }}",
                  "login": "{{ data.user.login }}"
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
  "resourceType": "issue",
  "id": 1606815439,
  "name": "Issue #6",
  "fullName": "mia-platform/eslint-config-mia#6",
  "repository": "eslint-config-mia",
  "organization": "mia-platform",
  "title": "Object destructuring on variable reassignment",
  "number": 6,
  "state": "open",
  "labels": [],
  "created_at": "2023-03-02T13:31:48Z",
  "updated_at": "2023-03-03T15:03:49Z",
  "user": {
    "id": 58828402,
    "login": "hiimjako"
  }
}
```
