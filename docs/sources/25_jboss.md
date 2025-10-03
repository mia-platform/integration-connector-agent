# JBoss/WildFly

The JBoss source allows the integration-connector-agent to monitor JBoss/WildFly application server deployments by polling the management API.

## Polling Integration

The JBoss source integrates with JBoss/WildFly by connecting to the management interface and periodically polling for deployment information. The following steps are performed:

1. **Authentication**: The source connects to the JBoss/WildFly management interface using HTTP Digest Authentication.
2. **Polling**: The source periodically queries the management API for deployment status information.
3. **Event Generation**: Deployment information is converted into events and sent to the configured pipelines.

### Service Configuration

The following configuration options are supported by the JBoss source:

- **type** (*string*): The type of the source, in this case `jboss`
- **wildflyUrl** (*string*) *optional*: The URL of the WildFly management interface. Defaults to `http://localhost:9990/management`.
- **username** (*string*) *optional*: The username for management interface authentication. Defaults to `admin`.
- **password** ([*SecretSource*](../20_install.md#secretsource)): The password for management interface authentication. **Required**.
- **pollingInterval** (*string*) *optional*: The interval between polling operations (e.g., "30s", "5m"). Defaults to `1s`.

#### Example Configuration

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

### How to Configure JBoss/WildFly

To configure JBoss/WildFly for monitoring, you need to:

1. **Enable Management Interface**: Ensure the management interface is enabled and accessible. By default, it runs on port 9990.

2. **Create Management User**: Create a management user with appropriate permissions:
   ```bash
   # For WildFly/JBoss EAP
   ./add-user.sh
   # Select 'Management User'
   # Enter username and password
   ```

3. **Configure Network Access**: If running the integration-connector-agent on a different host, ensure the management interface accepts remote connections:
   ```xml
   <!-- In standalone.xml or domain.xml -->
   <management-interfaces>
     <http-interface security-realm="ManagementRealm" http-upgrade-enabled="true">
       <socket-binding http="management-http"/>
     </http-interface>
   </management-interfaces>
   ```

4. **Test Connection**: Verify the management interface is accessible:
   ```bash
   curl -u username:password http://localhost:9990/management \
     -H "Content-Type: application/json" \
     -d '{"operation":"read-resource","address":["deployment","*"]}'
   ```

## Supported Events

The JBoss source generates the following event types:

### jboss:deployment_status

This event is generated for each deployment found in the JBoss/WildFly server.

#### Event Structure

```json
{
  "deployment": {
    "name": "wildfly-helloworld.war",
    "runtimeName": "wildfly-helloworld.war",
    "status": "OK",
    "enabled": true,
    "persistent": true,
    "content": [
      {
        "hash": {
          "BYTES_VALUE": "sivyKh3pYreM5wftC3jqVDLW2xc="
        }
      }
    ],
    "subdeployment": null,
    "subsystem": {
      "undertow": {
        "active-sessions": 0,
        "context-root": "/wildfly-helloworld",
        "server": "default-server",
        "sessions-created": 0,
        "virtual-host": "default-host",
        "servlet": {
          "org.jboss.as.quickstarts.helloworld.HelloWorldServlet": {
            "max-request-time": 0,
            "min-request-time": 0,
            "request-count": 0,
            "servlet-class": "org.jboss.as.quickstarts.helloworld.HelloWorldServlet",
            "servlet-name": "org.jboss.as.quickstarts.helloworld.HelloWorldServlet",
            "total-request-time": 0
          }
        }
      }
    }
  },
  "timestamp": "2025-10-03T10:30:00Z",
  "eventType": "jboss:deployment_status"
}
```

#### Available Template Fields

You can use the following fields in your processors and mappers:

##### Basic Deployment Information
- `{{ deployment.name }}` - The deployment name (e.g., "wildfly-helloworld.war")
- `{{ deployment.runtimeName }}` - The runtime name of the deployment
- `{{ deployment.status }}` - The deployment status (e.g., "OK", "FAILED")
- `{{ deployment.enabled }}` - Whether the deployment is enabled (boolean)
- `{{ deployment.persistent }}` - Whether the deployment is persistent (boolean)

##### Content Information
- `{{ deployment.content }}` - Array of content objects with hash information
- `{{ deployment.content[0].hash.BYTES_VALUE }}` - Content hash value

##### Subsystem Information (Undertow)
- `{{ deployment.subsystem.undertow.active-sessions }}` - Number of active HTTP sessions
- `{{ deployment.subsystem.undertow.context-root }}` - Application context root (e.g., "/wildfly-helloworld")
- `{{ deployment.subsystem.undertow.server }}` - Undertow server name (e.g., "default-server")
- `{{ deployment.subsystem.undertow.sessions-created }}` - Total sessions created
- `{{ deployment.subsystem.undertow.virtual-host }}` - Virtual host name (e.g., "default-host")

##### Servlet Information
- `{{ deployment.subsystem.undertow.servlet }}` - Complete servlet information object
- Individual servlet metrics (when accessing specific servlets by name)

##### Event Metadata
- `{{ timestamp }}` - Event generation timestamp
- `{{ eventType }}` - Always "jboss:deployment_status"

### Primary Keys

The JBoss source uses the following primary key for deployment events:
- `deploymentName`: The name of the deployment

## Authentication

The JBoss source uses **HTTP Digest Authentication** to connect to the JBoss/WildFly management interface. This is the standard authentication method for JBoss/WildFly management operations.

### Security Considerations

1. **Secure Passwords**: Store management passwords as environment variables or secure secrets.
2. **Network Security**: Ensure the management interface is only accessible from trusted networks.
3. **User Permissions**: Use dedicated management users with minimal required permissions.
4. **HTTPS**: Consider configuring HTTPS for the management interface in production environments.

## Troubleshooting

### Common Issues

1. **Connection Refused**: Verify the management interface is enabled and the URL is correct.
2. **Authentication Failed**: Check username and password credentials.
3. **No Deployments Found**: Ensure deployments are present and the user has read permissions.
4. **Network Timeout**: Adjust polling interval or check network connectivity.

### Debug Logging

Enable debug logging to troubleshoot issues:

```bash
LOG_LEVEL=debug ./integration-connector-agent
```

This will provide detailed information about:
- HTTP requests and responses
- Authentication process
- Deployment parsing
- Event generation

### Example Debug Output

```
DEBUG JBoss client: making initial HTTP request
DEBUG JBoss client: received response from management API
DEBUG JBoss client: parsed deployment with enhanced mapping
DEBUG Created deployment event, sending to pipeline/sink
```

## Performance Considerations

- **Polling Interval**: Balance between real-time updates and server load. Recommended: 10-60 seconds for production.
- **Network Latency**: Consider network latency when setting polling intervals.
- **Server Load**: Monitor JBoss/WildFly server performance when enabling monitoring.
- **Event Volume**: Large numbers of deployments will generate more events per polling cycle.
