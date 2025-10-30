# GCP Network mapping

This document describes the GCP Network mapping used to convert GCP inventory Pub/Sub events into a normalized asset event.

Purpose

- Normalize GCP Network events emitted by the inventory Pub/Sub source.
- Prepare a compact asset object with a consistent shape for downstream processing or sinks.

Mapped fields

- Id (resource ID)
- Name
- Description
- MTU
- Routing configuration (routingMode)
- Location

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
              "celExpression": "eventType == 'compute.googleapis.com/Network'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "id": "{{resource.data.id}}",
                "name": "{{resource.data.name}}",
                "description": "{{resource.data.description}}",
                "mtu": "{{resource.data.mtu}}",
                "routingConfig": "{{resource.data.routingConfig.routingMode}}",
                "location": "{{resource.location}}",
                "updateTime": "{{updateTime}}"
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
  "id": "447776895153723587",
  "name": "vpc-network-test",
  "description": "this is a test for a network",
  "mtu": 1460,
  "routingConfig": {
    "routingMode": "REGIONAL"
  },
  "updateTime": "2025-10-14T15:03:08.591868Z",
  "location": "global"
}
```
