# Microsoft Azure DevOps

The Microsoft Azure DevOps source allows the integration-connector-agent to receive events from Azure DevOps.

## Webhook Integration

The GitHub source integrates with webhooks by exposing an endpoint at `/azure-devops/webhook` (configurable).  
When a webhook wvent is received, the followint steps are performed:

1. **Validation**:  The request is validated using the secret passed by the Webhook.
1. **Repository load**: The source will get the list of all the repository availabe to the user and generate a new
  write event for every single one.

### Service Configuration

The following configuration options are supported by the GitHub source:

- **type** (*string*): The type of the source, in this case `azure-devops`
- **authentication** (*object*) *optional*: The authentication configuration
  - **secret** ([*SecretSource*](../20_install.md#secretsource)): The secret used to validate incoming webhook requests
- **webhookPath** (*string*) *optional*: The path where to receive the webhook events. Defaults to `/azure-devops/webhook`.
- **azureDevOpsOrganizationUrl** (*string*): The Azure DevOps organization URL
- **azureDevOpsPersonalAccessToken** ([*SecretSource*](../20_install.md#secretsource)): The PAT used to authorize the
  calls to th Azure DevOps endpoint

#### Example

```json
{
  "type": "azure-devops",
  "webhookPath": "/webhook",
  "authentication": {
    "secret": {
      "fromEnv": "SECRET"
    }
  },
  "azureDevOpsOrganizationUrl": "https://http://dev.azure.com/organizationName",
  "azureDevOpsPersonalAccessToken": {
    "fromEnv": "PAT_SECRET"
  }
}
```
