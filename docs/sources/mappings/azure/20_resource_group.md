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
		"tags": "{{tags}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name",
	"identity": null,
	"kind": "",
	"location": "northeurope",
	"managedBy": "",
	"name": "group-name",
	"plan": null,
	"properties": {
		"provisioningState": "Succeeded"
	},
	"resourceGroup": "group-name",
	"sku": null,
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.resources/subscriptions/resourcegroups",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/resources/resource-groups/get?view=rest-resources-2021-04-01#resourcegroup
