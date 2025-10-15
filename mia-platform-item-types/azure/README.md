# Azure Event Type Schemas

This directory contains JSON schemas for different Azure resource types that can be processed by Azure sources in the integration connector agent.

## Available Schemas

### Generic Resources

1. **azure-resource.json** - Generic Azure resource schema
   - Used for: All Azure resource types (virtual machines, networks, disks, Kubernetes clusters, etc.)
   - Contains: Common Azure resource metadata, location, relationships, tags, SKU, identity, and runtime data
   - Supports: Compute, Network, Storage, Container, Database, and other Azure services

### Specific Resources

1. **azure-virtual-network.json** - Azure Virtual Network resources (legacy - use azure-resource.json)
   - Used for: `microsoft.network/virtualnetworks` resource type
   - Contains: Virtual network metadata, location, relationships, tags, and runtime data

## Common Fields

All schemas include these standard fields:

- `name`: The Azure resource name
- `runtimeData`: Object containing detailed Azure resource information
  - `name`: Resource name
  - `type`: Azure resource type (e.g., "microsoft.compute/virtualmachines", "microsoft.network/virtualnetworks")
  - `provider`: Cloud provider identifier ("azure")
  - `location`: Azure region where the resource is deployed
  - `resourceGroup`: Azure resource group name
  - `subscriptionId`: Azure subscription ID (UUID format)
  - `relationships`: Array of relationship identifiers (subscription, resource group, dependencies)
  - `tags`: Azure resource tags as key-value pairs
  - `properties`: Resource-specific properties (varies by resource type)
  - `sku`: SKU information (name, tier, size, family, capacity)
  - `identity`: Managed identity information (system/user assigned)
  - `zones`: Availability zones for the resource
  - `kind`: Resource kind (used by some resource types)
  - `resourceId`: Full Azure resource ID
  - `provisioningState`: Resource provisioning state
  - `timestamp`: Data collection timestamp

## Supported Azure Resource Types

The generic `azure-resource.json` schema supports various Azure resource types including:

### Compute Resources
- Virtual Machines: `microsoft.compute/virtualmachines`
- Virtual Machine Scale Sets: `microsoft.compute/virtualmachinescalesets`
- Availability Sets: `microsoft.compute/availabilitysets`
- Disks: `microsoft.compute/disks`
- Snapshots: `microsoft.compute/snapshots`

### Network Resources
- Virtual Networks: `microsoft.network/virtualnetworks`
- Subnets: `microsoft.network/virtualnetworks/subnets`
- Network Security Groups: `microsoft.network/networksecuritygroups`
- Public IP Addresses: `microsoft.network/publicipaddresses`
- Load Balancers: `microsoft.network/loadbalancers`
- Application Gateways: `microsoft.network/applicationgateways`

### Container Resources
- Kubernetes Clusters: `microsoft.containerservice/managedclusters`
- Container Instances: `microsoft.containerinstance/containergroups`
- Container Registries: `microsoft.containerregistry/registries`

### Storage Resources
- Storage Accounts: `microsoft.storage/storageaccounts`
- Blob Services: `microsoft.storage/storageaccounts/blobservices`
- File Services: `microsoft.storage/storageaccounts/fileservices`

### Database Resources
- SQL Databases: `microsoft.sql/servers/databases`
- SQL Servers: `microsoft.sql/servers`
- Cosmos DB Accounts: `microsoft.documentdb/databaseaccounts`

### Other Resources
- Key Vaults: `microsoft.keyvault/vaults`
- App Services: `microsoft.web/sites`
- Function Apps: `microsoft.web/sites` (kind: "functionapp")
- Logic Apps: `microsoft.logic/workflows`

## Usage

These schemas can be used for:

1. **Data Validation**: Ensure incoming Azure data matches expected structure
2. **Console Catalog**: Define item type definitions for storing Azure entities
3. **Documentation**: Understanding the data structure for each resource type
4. **Mapping Configuration**: Reference for mapper processor field mappings

## Resource Type Mapping

The schemas correspond to Azure resource types:

- All Azure Resources → `azure-resource.json` (generic schema for any Azure resource type)
- Virtual Networks → `azure-virtual-network.json` (legacy - specific schema for microsoft.network/virtualnetworks)

Additional Azure resource types can be added as needed following the same pattern, or use the generic `azure-resource.json` schema.

## Schema Structure

Each Azure resource schema follows this general structure:

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Resource name"
    },
    "runtimeData": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "type": { "type": "string" },
        "provider": { "type": "string", "enum": ["azure"] },
        "location": { "type": "string" },
        "resourceGroup": { "type": "string" },
        "subscriptionId": { "type": "string" },
        "relationships": { "type": "array" },
        "tags": { "type": "object" },
        "properties": { "type": "object" },
        "sku": { "type": "object" },
        "identity": { "type": "object" },
        "timestamp": { "type": "string", "format": "date-time" }
      },
      "required": ["name", "type", "provider", "location", "timestamp"]
    }
  },
  "required": ["name", "runtimeData"]
}
```

## Example Data

Here are examples of how different Azure resources would map to the generic schema:

### Virtual Machine Example

```json
{
  "name": "my-vm",
  "runtimeData": {
    "name": "my-vm",
    "type": "microsoft.compute/virtualmachines",
    "provider": "azure",
    "location": "eastus",
    "resourceGroup": "my-rg",
    "subscriptionId": "12345678-1234-1234-1234-123456789012",
    "relationships": [
      "subscription/12345678-1234-1234-1234-123456789012",
      "resourceGroup/my-rg"
    ],
    "tags": {
      "environment": "production",
      "project": "web-app"
    },
    "properties": {
      "vmSize": "Standard_D2s_v3",
      "osType": "Linux",
      "provisioningState": "Succeeded"
    },
    "sku": {
      "name": "Standard_D2s_v3"
    },
    "zones": ["1"],
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```

### Kubernetes Cluster Example

```json
{
  "name": "my-aks",
  "runtimeData": {
    "name": "my-aks",
    "type": "microsoft.containerservice/managedclusters",
    "provider": "azure",
    "location": "westus2",
    "resourceGroup": "k8s-rg",
    "subscriptionId": "12345678-1234-1234-1234-123456789012",
    "relationships": [
      "subscription/12345678-1234-1234-1234-123456789012",
      "resourceGroup/k8s-rg"
    ],
    "tags": {
      "team": "platform",
      "cost-center": "engineering"
    },
    "properties": {
      "kubernetesVersion": "1.28.3",
      "nodeResourceGroup": "MC_k8s-rg_my-aks_westus2",
      "provisioningState": "Succeeded"
    },
    "identity": {
      "type": "SystemAssigned",
      "principalId": "abcd1234-ef56-gh78-ij90-klmnopqrstuv"
    },
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```
