# Pull Request

A GitHub Pull Request represents a proposed code change submitted to a repository.
Pull requests include a head and base reference, metadata, and author information.

The possible values a pull request can have are:

- resourceType: the resource type ("pull_request")
- id: unique numeric identifier for the pull request
- name: human-readable name
- fullName: repository full name with PR number
- repository: repository name
- organization: organization or owner name
- number: PR number within the repository
- title: PR title
- state: PR state (open, closed)
- createdAt: creation timestamp
- updatedAt: last update timestamp
- user.id: author user id
- user.login: author username

## Mapping Example

```json
{
  "integrations": [
    {
      "source": {
        "type": "github"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'github-import-pull_request'"
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
                "number": "{{ number }}",
                "title": "{{ data.title }}",
                "state": "{{ data.state }}",
                "createdAt": "{{ data.created_at }}",
                "updatedAt": "{{ data.updated_at }}",
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
  "resourceType": "pull_request",
  "id": 1262335574,
  "name": "PR #7",
  "fullName": "mia-platform/eslint-config-mia#7",
  "repository": "eslint-config-mia",
  "organization": "mia-platform",
  "number": 7,
  "title": "Allow variable reassignment without object destructuring",
  "state": "open",
  "createdAt": "2023-03-03T16:23:58Z",
  "updatedAt": "2023-03-08T09:17:27Z",
  "user": {
    "id": 58828402,
    "login": "hiimjako"
  }
}
```
