# Microsoft Azure Storage Account

For importing an Azure Storage Account in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"provisioningState": "{{properties.provisioningState}}",
		"allowBlobPublicAccess": "{{properties.allowBlobPublicAccess}}",
		"publicNetworkAccess": "{{properties.publicNetworkAccess}}",
		"isHnsEnabled": "{{properties.isHnsEnabled}}",
		"primaryLocation": "{{properties.primaryLocation}}",
		"statusOfPrimary": "{{properties.statusOfPrimary}}",
		"secondaryLocation": "{{properties.secondaryLocation}}",
		"statusOfSecondary": "{{properties.statusOfSecondary}}",
		"tags": "{{tags}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Storage/storageAccounts/account-name",
	"identity": {
		"type": "None"
	},
	"kind": "StorageV2",
	"location": "northeurope",
	"managedBy": "",
	"name": "account-name",
	"plan": null,
	"properties": {
		"accessTier": "Hot",
		"allowBlobPublicAccess": true,
		"allowCrossTenantReplication": true,
		"allowSharedKeyAccess": true,
		"creationTime": "1970-01-01T00:00:00.0000000Z",
		"defaultToOAuthAuthentication": false,
		"encryption": {
			"keySource": "Microsoft.Storage",
			"services": {
				"blob": {
					"enabled": true,
					"keyType": "Account",
					"lastEnabledTime": "1970-01-01T00:00:00.0000000Z"
				},
				"file": {
					"enabled": true,
					"keyType": "Account",
					"lastEnabledTime": "1970-01-01T00:00:00.0000000Z"
				}
			}
		},
		"isHnsEnabled": true,
		"isLocalUserEnabled": true,
		"isNfsV3Enabled": false,
		"isSftpEnabled": false,
		"keyCreationTime": {
			"key1": "1970-01-01T00:00:00.0000000Z",
			"key2": "1970-01-01T00:00:00.0000000Z"
		},
		"minimumTlsVersion": "TLS1_2",
		"networkAcls": {
			"bypass": "AzureServices",
			"defaultAction": "Allow",
			"ipRules": [],
			"ipv6Rules": [],
			"virtualNetworkRules": []
		},
		"primaryEndpoints": {
			"blob": "https://account-name.blob.core.windows.net/",
			"dfs": "https://account-name.dfs.core.windows.net/",
			"file": "https://account-name.file.core.windows.net/",
			"queue": "https://account-name.queue.core.windows.net/",
			"table": "https://account-name.table.core.windows.net/",
			"web": "https://account-name.z16.web.core.windows.net/"
		},
		"primaryLocation": "northeurope",
		"privateEndpointConnections": [],
		"provisioningState": "Succeeded",
		"publicNetworkAccess": "Enabled",
		"statusOfPrimary": "available",
		"supportsHttpsTrafficOnly": true
	},
	"resourceGroup": "group-name",
	"sku": {
		"name": "Standard_LRS",
		"tier": "Standard"
	},
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.storage/storageaccounts",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/storagerp/storage-accounts/get-properties?view=rest-storagerp-2024-01-01&tabs=HTTP#storageaccount
