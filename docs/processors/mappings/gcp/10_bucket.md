# GCP Bucket mapping

This document describes the GCP Bucket mapping used to convert GCP inventory Pub/Sub events into a normalized asset event.

Purpose

- Normalize GCP Storage Bucket events emitted by the inventory Pub/Sub source.
- Prepare a compact asset object with a consistent shape for downstream processing or sinks.

Mapped fields

- Id (resource ID)
- Name (bucket name)
- Labels
- Location
- LocationType
- StorageClass
- TimeCreated
- Updated timestamps
- Versioning
- HierarchicalNamespace

```json
{
  "integrations": [
    {
      "source": {
        "type": "gcp-inventory-pubsub"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "cloud-vendor-aggregator",
              "cloudVendorName": "gcp",
              "authOptions": {}
            },
            {
              "type": "filter",
              "celExpression": "eventType == 'storage.googleapis.com/Bucket'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "id": "{{resource.data.id}}",
                "name": "{{resource.data.name}}",
                "kind": "{{resource.data.kind}}",
                "labels": "{{resource.data.labels}}",
                "location": "{{resource.data.location}}",
                "locationType": "{{resource.data.locationType}}",
                "storageClass": "{{resource.data.storageClass}}",
                "timeCreated": "{{resource.data.timeCreated}}",
                "updated": "{{resource.data.updated}}",
                "versioning": "{{resource.data.versioning.enabled}}",
                "hierarchicalNamespace": "{{resource.data.hierarchicalNamespace.enabled}}"
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

## Example

```json
{
    "id": "custom-bucket-1470",
    "name": "custom-bucket-1470",
    "kind": "storage#bucket",
    "labels": {
        "custom": "1470"
    },
    "location": "US",
    "locationType": "multi-region",
    "storageClass": "STANDARD",
    "timeCreated": "2025-10-10T10:38:12.324Z",
    "updated": "2025-10-10T10:38:12.324Z",
    "versioning": false,
    "hierarchicalNamespace": false,
}
```
