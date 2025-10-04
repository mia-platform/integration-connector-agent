# JBoss/WildFly Source

This source monitors JBoss/WildFly deployments via the management API using HTTP digest authentication. It polls the management interface at configurable intervals to retrieve deployment status information.

## Configuration

```json
{
  "type": "jboss",
  "wildflyUrl": "http://localhost:9990/management",
  "username": "admin",
  "password": "your-password",
  "pollingInterval": "30s"
}
```

### Configuration Options

- `wildflyUrl` (optional): WildFly management URL. Default: `http://localhost:9990/management`
- `username` (optional): Management username. Default: `admin`
- `password` (required): Management password
- `pollingInterval` (optional): Polling interval for checking deployments. Default: `30s`

## Features

- **Digest Authentication**: Supports HTTP digest authentication with WildFly management interface
- **Deployment Monitoring**: Monitors all deployments and their status
- **Configurable Polling**: Adjustable polling interval for monitoring frequency
- **Event Generation**: Creates pipeline events for deployment status changes

## Events

The source generates events of type `jboss:deployment_status` with the following structure:

```json
{
  "deployment": {
    "name": "my-app.war",
    "runtimeName": "my-app.war", 
    "status": "OK",
    "enabled": true,
    "persistentDeployed": true
  },
  "timestamp": "2025-01-01T12:00:00Z",
  "eventType": "jboss:deployment_status"
}
```

## Requirements

- JBoss/WildFly with management interface enabled
- Management user with appropriate permissions
- Network connectivity to the management interface
