# integration-connector-agent

## Development Local

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
