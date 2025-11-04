# Repository

A GitHub Repository resource represents a code repository hosted on GitHub.
Repositories include metadata like name, description, owner, visibility, and timestamps.

The possible values a repository can have are:

- resourceType: the resource type ("repository")
- id: unique numeric identifier for the repository
- name: repository name
- fullName: full repository name (owner/name)
- organization: owning organization
- description: repository description
- defaultBranch: default branch name
- language: primary language
- cloneUrl: HTTPS clone URL
- private: whether the repository is private
- visibility: repository visibility (public, private)
- createdAt: creation timestamp
- updatedAt: last update timestamp
- pushedAt: last push timestamp
- owner.id: owner id
- owner.login: owner login
- owner.type: owner type (User or Organization)
- licenseName: license name if present

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
              "celExpression": "eventType == 'repository'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "resourceType": "{{ type }}",
                "id": "{{ id }}",
                "name": "{{ name }}",
                "fullName": "{{ full_name }}",
                "organization": "{{ organization }}",
                "description": "{{ data.description }}",
                "defaultBranch": "{{ data.default_branch }}",
                "language": "{{ data.language }}",
                "cloneUrl": "{{ data.clone_url }}",
                "private": "{{ data.private }}",
                "visibility": "{{ data.visibility }}",
                "createdAt": "{{ data.created_at }}",
                "updatedAt": "{{ data.updated_at }}",
                "pushedAt": "{{ data.pushed_at }}",
                "owner": {
                  "id": "{{ data.owner.id }}",
                  "login": "{{ data.owner.login }}",
                  "type": "{{ data.owner.type }}"
                },
                "licenseName": "{{ data.license.name }}"
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
  "resourceType": "repository",
  "id": 1081025633,
  "name": "crud-service-universal",
  "fullName": "mia-platform/crud-service-universal",
  "organization": "mia-platform",
  "description": "A high-performance CRUD service written in Rust for both SQL and NoSQL DBs",
  "defaultBranch": "main",
  "language": "Rust",
  "cloneUrl": "https://github.com/mia-platform/crud-service-universal.git",
  "private": false,
  "visibility": "public",
  "createdAt": "2025-10-22T07:50:59Z",
  "updatedAt": "2025-10-22T07:53:39Z",
  "pushedAt": "2025-10-22T07:53:34Z",
  "owner": {
    "id": 10514842,
    "login": "mia-platform",
    "type": "Organization"
  },
  "licenseName": ""
}
```
