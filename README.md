# Integration Connector Agent

The Integration Connector Agent is a powerful data synchronization tool that connects external sources with multiple sinks,
enabling real-time data flow and transformation between different systems. It's designed to simplify data integration
workflows by providing a flexible, pipeline-based architecture.

## 🚀 Key Features

- **Multi-Source Support**: Connect to various external data sources (GitHub, GitLab, Jira, Confluence, Azure, AWS,
  GCP, and more)
- **Flexible Data Processing**: Transform data through configurable processor pipelines (Filter, Mapper, RPC Plugin,
  Cloud Vendor Aggregator)
- **Multiple Sink Options**: Send data to different destinations (Mia-Platform Console Catalog, MongoDB, CRUD Service, Kafka)
- **Real-time Synchronization**: Keep data synchronized between sources and sinks with minimal latency
- **Cloud-Native**: Docker-ready with Kubernetes support for scalable deployments

## 📚 Documentation

For comprehensive documentation, examples, and configuration guides:

### Getting Started

- [📖 Overview & Features](./docs/10_overview.md) - Complete feature overview and supported integrations
- [⚙️ Installation Guide](./docs/20_install.md) - Installation and setup instructions
- [🏗️ Architecture](./docs/30_architecture.md) - System architecture and data flow diagrams

### Configuration Guides

- [🔌 Sources Documentation](./docs/sources/) - All available source integrations
- [📤 Sinks Documentation](./docs/sinks/) - All available sink destinations  
- [⚡ Processors Documentation](./docs/processors/) - Data transformation processors

### Examples & Schemas

- [📋 Configuration Examples](./examples/) - Ready-to-use configuration files
- [🏷️ Item Type Schemas](./mia-platform-item-types/) - JSON schemas for Mia-Platform Console Catalog

### Troubleshooting

- [🛠️ Troubleshooting Guide](./docs/troubleshooting/) - Common issues and solutions
- [📝 Migration Guides](./docs/GITHUB_AUTHENTICATION_MIGRATION.md) - Upgrade instructions

## 🎯 Quick Start Example

Here's a simple example that integrates JBoss deployments with the Mia-Platform Console Catalog:

```json
{
  "integrations": [
    {
      "source": {
        "type": "jboss",
        "wildflyUrl": "http://localhost:9990/management",
        "username": "admin",
        "password": { "fromEnv": "JBOSS_PASSWORD" },
        "pollingInterval": "3s"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'jboss:deployment_status'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "deploymentName": "{{ deployment.name }}",
                "status": "{{ deployment.status }}",
                "timestamp": "{{ timestamp }}"
              }
            }
          ],
          "sinks": [
            {
              "type": "console-catalog",
              "baseUrl": "https://your-console-url.com",
              "itemType": "jboss-application"
            }
          ]
        }
      ]
    }
  ]
}
```

For more examples and complete configuration options, see the [documentation](./docs/) and [examples](./examples/) directories.

## 🔧 Development Local

To develop the service locally you need:

- Go 1.24+

To start the application locally

create a config.json like (for JBoss integration, see more in [/docs])
or copy from `cp config.json.local config.json`

```json
{
  "integrations": [
    {
      "source": {
        "type": "jboss",
        "wildflyUrl": "http://localhost:9990/management",
        "username": "admin",
        "password": {
          "fromEnv": "JBOSS_PASSWORD"
        },
        "pollingInterval": "3s"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'jboss:deployment_status'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "deploymentName": "{{ deployment.name }}",
                "status": "{{ deployment.status }}",
                "enabled": "{{ deployment.enabled }}",
                "runtimeName": "{{ deployment.runtimeName }}",
                "persistent": "{{ deployment.persistent }}",
                "content": "{{ deployment.content }}",
                "subsystem": "{{ deployment.subsystem }}",
                "timestamp": "{{ timestamp }}",
                "source": "jboss"
              }
            }
          ],
          "sinks": [
            {
              "type": "console-catalog",
              "url": "https://your-console-url.com",
              "tenantId": "your-tenant-id",
              "clientId": "your-client-id",
              "clientSecret": { "fromEnv": "CONSOLE_CLIENT_SECRET" },
              "itemTypeDefinitionRef": {
                "name": "jboss-application",
                "namespace": "your-tenant-id"
              },
              "itemNameTemplate": "{{deploymentName}} - {{runtimeName}}"
            }
          ]
        }
      ]
    }
  ]
}
```

Build the code

```bash
go build .
```

Run the application. This is an example with JBoss integration. More details in [/docs]

```bash
CONFIGURATION_PATH=./config.json JBOSS_PASSWORD='your-jboss-pwd' LOG_LEVEL=debug CONSOLE_TENANT_ID=your-mia-tenat-id CONSOLE_SERVICE_ACCOUNT_CLIENT_ID=mia-client-id CONSOLE_SERVICE_ACCOUNT_CLIENT_SECRET=mia-client-secret ./integration-connector-agent
```

If you want to echo the pipeline use this config instead

```json
{
  "integrations": [
    {
      "source": {
        "type": "jboss",
        "wildflyUrl": "http://localhost:9990/management",
        "username": "admin",
        "password": {
          "fromEnv": "JBOSS_PASSWORD"
        },
        "pollingInterval": "3s"
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "filter",
              "celExpression": "eventType == 'jboss:deployment_status'"
            },
            {
              "type": "mapper",
              "outputEvent": {
                "deploymentName": "{{ deployment.name }}",
                "status": "{{ deployment.status }}",
                "enabled": "{{ deployment.enabled }}",
                "runtimeName": "{{ deployment.runtimeName }}",
                "persistent": "{{ deployment.persistent }}",
                "content": "{{ deployment.content }}",
                "subsystem": "{{ deployment.subsystem }}",
                "timestamp": "{{ timestamp }}",
                "source": "jboss"
              }
            }
          ],
          "sinks": [
            {
              "type": "fake"
            }
          ]
        }
      ]
    }
  ]
}
```

By default the service will run on port 8080, to change the port please set `HTTP_PORT` env variable

## Testing

To test the application use:

```go
make test
```

## License

`integration-connector-agent` is licensed under [AGPL-3.0-only](./LICENSE). For Commercial and other
exceptions please read [LICENSING.md](./LICENSING.md)
