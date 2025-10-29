# Microsoft Azure Subscription

For importing an Azure Subscription in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"displayName": "{{name}}",
		"state": "{{properties.state}}",
		"tags": "{{tags}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"identity": null,
	"kind": "",
	"location": "",
	"managedBy": "",
	"name": "Microsoft Azure Sponsorship",
	"plan": null,
	"properties": {
		"managedByTenants": [
			{
				"tenantId": "3c3c3c3c-3c3c-3c3c-3c3c-3c3c3c3c3c3c"
			},
		],
		"managementGroupAncestorsChain": [
			{
				"displayName": "Tenant Root Group",
				"name": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a"
			}
		],
		"state": "Enabled",
		"subscriptionPolicies": {
			"locationPlacementId": "Public_2014-09-01",
			"quotaId": "Sponsored_2016-01-01",
			"spendingLimit": "Off"
		}
	},
	"resourceGroup": "",
	"sku": null,
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": null,
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.resources/subscriptions",
	"zones": null
}
```
