# Azure Resource Examples

This file contains examples of how different Azure resources map to the generic `azure-resource.json` schema.

## Virtual Network Example

Based on the original JSON structure:

```json
{
  "name": "tools",
  "runtimeData": {
    "name": "tools",
    "type": "microsoft.network/virtualnetworks",
    "provider": "azure",
    "location": "northeurope",
    "resourceGroup": "tools",
    "subscriptionId": "633fa2a8-982c-431d-9535-a90d8948b27c",
    "relationships": [
      "subscription/633fa2a8-982c-431d-9535-a90d8948b27c",
      "resourceGroup/tools"
    ],
    "tags": {},
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```

## Additional Resource Examples

### Virtual Machine

```json
{
  "name": "web-server-01",
  "runtimeData": {
    "name": "web-server-01",
    "type": "microsoft.compute/virtualmachines",
    "provider": "azure",
    "location": "eastus",
    "resourceGroup": "production-rg",
    "subscriptionId": "633fa2a8-982c-431d-9535-a90d8948b27c",
    "relationships": [
      "subscription/633fa2a8-982c-431d-9535-a90d8948b27c",
      "resourceGroup/production-rg",
      "virtualNetwork/prod-vnet",
      "subnet/web-subnet"
    ],
    "tags": {
      "environment": "production",
      "project": "web-application",
      "owner": "platform-team"
    },
    "properties": {
      "vmSize": "Standard_D2s_v3",
      "osType": "Linux",
      "imageReference": {
        "publisher": "Canonical",
        "offer": "0001-com-ubuntu-server-focal",
        "sku": "20_04-lts-gen2",
        "version": "latest"
      }
    },
    "sku": {
      "name": "Standard_D2s_v3"
    },
    "zones": ["1"],
    "provisioningState": "Succeeded",
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```

### Kubernetes Cluster

```json
{
  "name": "production-aks",
  "runtimeData": {
    "name": "production-aks",
    "type": "microsoft.containerservice/managedclusters",
    "provider": "azure",
    "location": "westeurope",
    "resourceGroup": "k8s-production",
    "subscriptionId": "633fa2a8-982c-431d-9535-a90d8948b27c",
    "relationships": [
      "subscription/633fa2a8-982c-431d-9535-a90d8948b27c",
      "resourceGroup/k8s-production",
      "virtualNetwork/aks-vnet"
    ],
    "tags": {
      "environment": "production",
      "team": "platform",
      "cost-center": "engineering"
    },
    "properties": {
      "kubernetesVersion": "1.28.3",
      "nodeResourceGroup": "MC_k8s-production_production-aks_westeurope",
      "dnsPrefix": "production-aks-dns",
      "agentPoolProfiles": [
        {
          "name": "nodepool1",
          "count": 3,
          "vmSize": "Standard_D4s_v3",
          "osType": "Linux"
        }
      ]
    },
    "identity": {
      "type": "SystemAssigned",
      "principalId": "12345678-abcd-efgh-ijkl-123456789012"
    },
    "provisioningState": "Succeeded",
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```

### Storage Account

```json
{
  "name": "prodstorageacct001",
  "runtimeData": {
    "name": "prodstorageacct001",
    "type": "microsoft.storage/storageaccounts",
    "provider": "azure",
    "location": "westus2",
    "resourceGroup": "storage-rg",
    "subscriptionId": "633fa2a8-982c-431d-9535-a90d8948b27c",
    "relationships": [
      "subscription/633fa2a8-982c-431d-9535-a90d8948b27c",
      "resourceGroup/storage-rg"
    ],
    "tags": {
      "environment": "production",
      "backup": "enabled",
      "retention": "7-years"
    },
    "properties": {
      "accountType": "Standard_LRS",
      "encryption": {
        "services": {
          "blob": {
            "enabled": true
          },
          "file": {
            "enabled": true
          }
        }
      },
      "accessTier": "Hot"
    },
    "sku": {
      "name": "Standard_LRS",
      "tier": "Standard"
    },
    "kind": "StorageV2",
    "provisioningState": "Succeeded",
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```

### Managed Disk

```json
{
  "name": "web-server-01-disk",
  "runtimeData": {
    "name": "web-server-01-disk",
    "type": "microsoft.compute/disks",
    "provider": "azure",
    "location": "eastus",
    "resourceGroup": "production-rg",
    "subscriptionId": "633fa2a8-982c-431d-9535-a90d8948b27c",
    "relationships": [
      "subscription/633fa2a8-982c-431d-9535-a90d8948b27c",
      "resourceGroup/production-rg",
      "virtualMachine/web-server-01"
    ],
    "tags": {
      "environment": "production",
      "backup": "daily"
    },
    "properties": {
      "diskSizeGB": 128,
      "diskState": "Attached",
      "osType": "Linux",
      "creationData": {
        "createOption": "FromImage"
      }
    },
    "sku": {
      "name": "Premium_LRS",
      "tier": "Premium"
    },
    "zones": ["1"],
    "managedBy": "/subscriptions/633fa2a8-982c-431d-9535-a90d8948b27c/resourceGroups/production-rg/providers/Microsoft.Compute/virtualMachines/web-server-01",
    "provisioningState": "Succeeded",
    "timestamp": "2025-10-14T08:33:57.165510813Z"
  }
}
```
