# Azure Pipeline Configurations

This directory contains three different Azure pipeline configurations for integrating Azure resources with the Mia Platform Console Catalog.

## Configuration Files

### 1. azure-complete-pipeline-config.json
**Complete Configuration with All Fields**

This configuration maps ALL available Azure fields from the cloud vendor aggregator output to the Console Catalog. It includes:

- **All runtime data fields**: name, type, provider, location, resourceGroup, subscriptionId, relationships, tags, properties
- **SKU information**: name, tier, size, family, capacity
- **Identity management**: type, principalId, tenantId, userAssignedIdentities  
- **Additional metadata**: zones, kind, resourceId, etag, managedBy, plan, provisioningState, timestamps
- **Rich catalog metadata**: comprehensive labels, annotations, links, and Azure Portal integration

**Use this when**: You need complete Azure resource information in your catalog and have complex governance/compliance requirements.

### 2. azure-simple-pipeline-config.json
**Simplified Configuration for Core Fields Only**

This configuration focuses on the most commonly used Azure fields for better performance:

- **Essential fields**: name, type, provider, location, resourceGroup, subscriptionId
- **Basic metadata**: relationships, tags, properties, provisioningState, timestamp
- **Simplified catalog metadata**: basic labels and annotations
- **Default fallback values** for missing fields

**Use this when**: You want good performance and only need core Azure resource information.

### 3. azure-conditional-pipeline-config.json
**Balanced Configuration with Optional Fields**

This configuration includes all fields but handles optional ones gracefully:

- **All Azure fields** mapped
- **Standard JSON format** (no conditional templating syntax)
- **Comprehensive catalog metadata** with all Azure fields
- **Rich Azure Portal integration** with proper links

**Use this when**: You want comprehensive data but need to handle resources that may not have all optional fields populated.

## Field Mapping Reference

### Core Azure Fields (Available in all configurations)
```json
{
  "name": "{{name}}",                    // Resource name
  "runtimeData": {
    "name": "{{name}}",                  // Resource name  
    "type": "{{type}}",                  // Azure resource type (e.g., microsoft.compute/virtualmachines)
    "provider": "{{provider}}",          // Always "azure"
    "location": "{{location}}",          // Azure region (e.g., "eastus", "westeurope")
    "relationships": "{{relationships}}", // Array of parent/child relationships
    "tags": "{{tags}}",                  // Azure resource tags
    "timestamp": "{{timestamp}}"         // Data collection timestamp
  }
}
```

### Extended Azure Fields (Complete configuration only)
```json
{
  "resourceGroup": "{{resourceGroup}}",           // Resource group name
  "subscriptionId": "{{subscriptionId}}",         // Azure subscription ID
  "properties": "{{properties}}",                 // Resource-specific properties
  "sku": {                                        // SKU information
    "name": "{{sku.name}}",
    "tier": "{{sku.tier}}", 
    "size": "{{sku.size}}",
    "family": "{{sku.family}}",
    "capacity": "{{sku.capacity}}"
  },
  "identity": {                                   // Managed identity
    "type": "{{identity.type}}",
    "principalId": "{{identity.principalId}}",
    "tenantId": "{{identity.tenantId}}",
    "userAssignedIdentities": "{{identity.userAssignedIdentities}}"
  },
  "zones": "{{zones}}",                           // Availability zones
  "kind": "{{kind}}",                             // Resource kind  
  "resourceId": "{{resourceId}}",                 // Full Azure resource ID
  "etag": "{{etag}}",                             // Resource etag
  "managedBy": "{{managedBy}}",                   // Managing resource
  "plan": {                                       // Marketplace plan
    "name": "{{plan.name}}",
    "publisher": "{{plan.publisher}}",
    "product": "{{plan.product}}",
    "promotionCode": "{{plan.promotionCode}}",
    "version": "{{plan.version}}"
  },
  "provisioningState": "{{provisioningState}}",   // Resource state
  "createdTime": "{{createdTime}}",               // Creation timestamp
  "changedTime": "{{changedTime}}"                // Last modified timestamp
}
```

## Catalog Metadata Features

### Labels
Used for filtering and searching in the Console Catalog:
- `provider`: Cloud provider ("azure")
- `location`: Azure region
- `resourceType`: Azure resource type
- `resourceGroup`: Resource group name
- `provisioningState`: Resource state
- `skuName`/`skuTier`: SKU information

### Annotations
Used for detailed metadata and integration:
- `azure.com/*`: All Azure-specific metadata
- `azure.com/resource-id`: Full Azure resource ID for direct API access
- `azure.com/subscription-id`: Subscription for billing and access control
- `azure.com/import-source`: Data source tracking

### Links
Direct integration with Azure Portal:
- **Azure Portal**: Direct link to resource in Azure Portal
- **Resource Group**: Link to resource group management
- **Subscription**: Link to subscription overview

## Authentication Configuration

All configurations require these environment variables:

```bash
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id  
AZURE_CLIENT_SECRET=your-client-secret
CONSOLE_SERVICE_ACCOUNT_CLIENT_SECRET=your-console-secret
```

## Usage Examples

### Complete Configuration
```bash
# Use when you need all Azure metadata
cp azure-complete-pipeline-config.json your-integration-config.json
```

### Simple Configuration  
```bash
# Use for better performance with core fields only
cp azure-simple-pipeline-config.json your-integration-config.json
```

### Balanced Configuration
```bash
# Use for comprehensive data with graceful handling of optional fields
cp azure-conditional-pipeline-config.json your-integration-config.json
```

## Azure Resource Types Supported

The configurations work with all Azure resource types including:

- **Compute**: Virtual Machines, Scale Sets, Availability Sets, Disks
- **Network**: Virtual Networks, Subnets, Load Balancers, Application Gateways  
- **Storage**: Storage Accounts, Blob Services, File Services
- **Container**: AKS Clusters, Container Instances, Container Registries
- **Database**: SQL Databases, Cosmos DB, MySQL, PostgreSQL
- **Security**: Key Vaults, Security Centers
- **Web**: App Services, Function Apps, Logic Apps
- **And many more**: The generic schema supports any Azure resource type

## Performance Considerations

- **Complete Configuration**: Highest data fidelity, more processing overhead
- **Simple Configuration**: Best performance, essential data only  
- **Balanced Configuration**: Good balance of data completeness and performance

Choose based on your specific needs for data completeness vs. processing performance.