# Azure Permissions Troubleshooting Guide

This guide helps you resolve common Azure permission issues when using the Azure Activity Log Event Hub source.

## Common Permission Errors

### 1. Storage Blob Permission Error (403 AuthorizationPermissionMismatch)

**Error Message:**
```
GET https://[storage-account].blob.core.windows.net/[container]
RESPONSE 403: 403 This request is not authorized to perform this operation using this permission.
ERROR CODE: AuthorizationPermissionMismatch
```

**Root Cause:** The service principal lacks required permissions on the blob storage account used for Event Hub checkpoints.

**Solution:**
1. Go to Azure Portal → Storage Accounts → [your-storage-account] → Access Control (IAM)
2. Click "Add role assignment"
3. Select **Storage Blob Data Contributor** role
4. Assign to your service principal or managed identity

**Required Permissions:**
- `Storage Blob Data Contributor` (recommended)
- OR both `Storage Blob Data Reader` + `Storage Blob Data Writer`

### 2. Event Hub Access Error (403 Forbidden)

**Error Message:**
```
failed to create Event Hub consumer client for '[namespace]/[hub]'
403 Forbidden
```

**Root Cause:** The service principal lacks Event Hub access permissions.

**Solution:**
1. Go to Azure Portal → Event Hubs → [your-namespace] → Access Control (IAM)
2. Click "Add role assignment"
3. Select **Azure Event Hubs Data Receiver** role
4. Assign to your service principal or managed identity

**Required Permissions:**
- `Azure Event Hubs Data Receiver` (minimum)
- `Azure Event Hubs Data Owner` (for administrative operations)

### 3. Authentication Errors (401 Unauthorized)

**Error Message:**
```
failed to create Azure credentials
401 Unauthorized
```

**Root Cause:** Invalid or expired Azure credentials.

**Solution:**
1. Verify your Azure credentials:
   - `AZURE_TENANT_ID`: Correct tenant ID
   - `AZURE_CLIENT_ID`: Valid service principal client ID
   - `AZURE_CLIENT_SECRET`: Valid and not expired client secret

2. Check service principal status:
   ```bash
   az ad sp show --id <client-id>
   ```

3. Verify client secret expiration:
   ```bash
   az ad sp credential list --id <client-id>
   ```

## Complete Permission Setup Checklist

### Service Principal Setup
1. **Create Service Principal:**
   ```bash
   az ad sp create-for-rbac --name "integration-connector-agent" \
     --role "Reader" \
     --scopes "/subscriptions/<subscription-id>"
   ```

### Event Hub Permissions
2. **Assign Event Hub Data Receiver:**
   ```bash
   az role assignment create \
     --assignee <service-principal-id> \
     --role "Azure Event Hubs Data Receiver" \
     --scope "/subscriptions/<subscription-id>/resourceGroups/<rg>/providers/Microsoft.EventHub/namespaces/<namespace>"
   ```

### Storage Account Permissions
3. **Assign Storage Blob Data Contributor:**
   ```bash
   az role assignment create \
     --assignee <service-principal-id> \
     --role "Storage Blob Data Contributor" \
     --scope "/subscriptions/<subscription-id>/resourceGroups/<rg>/providers/Microsoft.Storage/storageAccounts/<storage-account>"
   ```

### Subscription Permissions (for webhook import)
4. **Assign Reader role on subscription:**
   ```bash
   az role assignment create \
     --assignee <service-principal-id> \
     --role "Reader" \
     --scope "/subscriptions/<subscription-id>"
   ```

## Verification Commands

### Test Event Hub Access
```bash
# Test Event Hub connection
az eventhubs eventhub show \
  --namespace-name <namespace> \
  --name <event-hub> \
  --resource-group <resource-group>
```

### Test Storage Account Access
```bash
# Test storage account access
az storage container show \
  --name <container> \
  --account-name <storage-account>
```

### Check Role Assignments
```bash
# List all role assignments for service principal
az role assignment list --assignee <service-principal-id>
```

## Configuration Example

```json
{
  "type": "azure-activity-log-event-hub",
  "subscriptionId": "00000000-0000-0000-0000-000000000000",
  "eventHubNamespace": "my-namespace",
  "eventHubName": "activity-log-hub",
  "checkpointStorageAccountName": "checkpointstorage",
  "checkpointStorageContainerName": "checkpoints",
  "tenantId": "tenant-id",
  "clientId": {
    "fromEnv": "AZURE_CLIENT_ID"
  },
  "clientSecret": {
    "fromEnv": "AZURE_CLIENT_SECRET"
  }
}
```

## Environment Variables

```bash
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
```

## Network Considerations

If you're running in a restricted network environment:

1. **Service Tags:** Ensure outbound access to `AzureEventHub` and `Storage` service tags
2. **Private Endpoints:** Configure private endpoints if using VNet integration
3. **DNS Resolution:** Verify DNS resolution for `*.servicebus.windows.net` and `*.blob.core.windows.net`

## Support Resources

- [Azure Event Hubs Documentation](https://docs.microsoft.com/en-us/azure/event-hubs/)
- [Azure Storage Access Control](https://docs.microsoft.com/en-us/azure/storage/common/storage-auth)
- [Azure RBAC Documentation](https://docs.microsoft.com/en-us/azure/role-based-access-control/)