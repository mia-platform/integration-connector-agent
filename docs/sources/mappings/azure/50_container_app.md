# Microsoft Azure Container App

For importing an Azure Container App in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"provisioningState": "{{properties.provisioningState}}",
		"tags": "{{tags}}",
		"runningStatus": "{{properties.runningStatus}}",
		"fqdn": "{{properties.configuration.ingress.fqdn}}",
		"external": "{{properties.configuration.ingress.external}}",
		"targetPort": "{{properties.configuration.ingress.targetPort}}",
		"minReplicas": "{{properties.template.scale.minReplicas}}",
		"maxReplicas": "{{properties.template.scale.maxReplicas}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.App/containerApps/app-name",
	"identity": {
		"type": "None"
	},
	"kind": "containerapps",
	"location": "westeurope",
	"managedBy": "",
	"name": "app-name",
	"plan": null,
	"properties": {
		"configuration": {
			"activeRevisionsMode": "Single",
			"dapr": null,
			"identitySettings": [],
			"ingress": null,
			"maxInactiveRevisions": 100,
			"registries": null,
			"runtime": null,
			"secrets": null,
			"service": null
		},
		"customDomainVerificationId": "D4E46985E8E1E3D6709EF3EF4B12C489DB0B0AECB10EE08582E55856E55D2B6C",
		"delegatedIdentities": [],
		"environmentId": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.App/managedEnvironments/managedEnvironment-group-name-0000",
		"eventStreamEndpoint": "https://westeurope.azurecontainerapps.dev/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/containerApps/app-name/eventstream",
		"latestReadyRevisionName": "app-name--0000000",
		"latestRevisionFqdn": "",
		"latestRevisionName": "app-name--0000000",
		"managedEnvironmentId": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.App/managedEnvironments/managedEnvironment-group-name-0000",
		"outboundIpAddresses": [],
		"patchingMode": "Automatic",
		"provisioningState": "Succeeded",
		"runningStatus": "Running",
		"template": {
			"containers": [
				{
					"image": "docker.io/nginx:latest",
					"imageType": "ContainerImage",
					"name": "app-name",
					"resources": {
						"cpu": 0.5,
						"ephemeralStorage": "2Gi",
						"memory": "1Gi"
					}
				}
			],
			"initContainers": null,
			"revisionSuffix": "",
			"scale": {
				"maxReplicas": 1,
				"minReplicas": 1,
				"rules": null
			},
			"serviceBinds": null,
			"terminationGracePeriodSeconds": null,
			"volumes": null
		},
		"workloadProfileName": "Consumption"
	},
	"resourceGroup": "group-name",
	"sku": null,
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {
		"test": ""
	},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.app/containerapps",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/resource-manager/containerapps/container-apps/get?view=rest-resource-manager-containerapps-2025-07-01&tabs=HTTP#containerapp
