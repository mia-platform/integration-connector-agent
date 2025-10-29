# Microsoft Azure Resource Group

For importing an Azure Resource Group in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"provisioningState": "{{properties.provisioningState}}",
		"tags": "{{tags}}",
		"addressPrefixes": "{{properties.addressSpace.addressPrefixes}}",
		"subnets": "{{properties.subnets}}
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/virtualNetworks/network-name",
	"identity": null,
	"kind": "",
	"location": "northeurope",
	"managedBy": "",
	"name": "network-name",
	"plan": null,
	"properties": {
		"addressSpace": {
			"addressPrefixes": [
				"10.1.0.0/24"
			]
		},
		"enableDdosProtection": false,
		"privateEndpointVNetPolicies": "Disabled",
		"provisioningState": "Succeeded",
		"resourceGuid": "5g5g5g5g-5g5g-5g5g-5g5g-5g5g5g5g5g5g",
		"subnets": [
			{
				"etag": "W/\"6d6d6d6d-6d6d-6d6d-6d6d-6d6d6d6d6d6d\"",
				"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/virtualNetworks/network-name/subnets/default",
				"name": "default",
				"properties": {
					"addressPrefix": "10.1.0.0/24",
					"delegations": [],
					"privateEndpointNetworkPolicies": "Disabled",
					"privateLinkServiceNetworkPolicies": "Enabled",
					"provisioningState": "Succeeded",
					"serviceAssociationLinks": [],
					"serviceEndpoints": [
						{
							"locations": [
								"northeurope",
								"westeurope"
							],
							"provisioningState": "Succeeded",
							"service": "Microsoft.Storage"
						}
					]
				},
				"type": "Microsoft.Network/virtualNetworks/subnets"
			}
		],
		"virtualNetworkPeerings": []
	},
	"resourceGroup": "azure-operationslab",
	"sku": null,
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.network/virtualnetworks",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/get?view=rest-virtualnetwork-2024-10-01&tabs=HTTP#virtualnetwork
