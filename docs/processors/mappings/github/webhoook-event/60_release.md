# Release

A GitHub Release event on a repository hosted on GitHub.

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
              "celExpression": "eventType == 'release'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "eventType": "{{ eventType }}",
                "action": "{{ action }}",
                "releaseId": "{{ release.id }}",
                "releaseTagName": "{{ release.tag_name }}",
                "releaseName": "{{ release.name }}",
                "releaseBody": "{{ release.body }}",
                "releaseHtmlUrl": "{{ release.html_url }}",
                "releaseUrl": "{{ release.url }}",
                "releaseAssetsUrl": "{{ release.assets_url }}",
                "releaseTarballUrl": "{{ release.tarball_url }}",
                "releaseZipballUrl": "{{ release.zipball_url }}",
                "releaseUploadUrl": "{{ release.upload_url }}",
                "releaseTargetCommitish": "{{ release.target_commitish }}",
                "releaseDraft": "{{ release.draft }}",
                "releasePrerelease": "{{ release.prerelease }}",
                "releaseMakeLatest": "{{ release.make_latest }}",
                "releaseGenerateReleaseNotes": "{{ release.generate_release_notes }}",
                "releaseCreatedAt": "{{ release.created_at }}",
                "releasePublishedAt": "{{ release.published_at }}",
                "releaseAuthorLogin": "{{ release.author.login }}",
                "releaseAuthorId": "{{ release.author.id }}",
                "releaseAuthorType": "{{ release.author.type }}",
                "releaseAssets": "{{ release.assets }}",
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
