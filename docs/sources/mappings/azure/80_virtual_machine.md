# Microsoft Azure Virtual Machine

For importing an Azure Virtual Machine in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"size": "{{properties.hardwareProfile.vmSize}}",
		"provisioningState": "{{properties.provisioningState}}",
		"osType": "{{properties.storageProfile.osDisk.osType}}",
		"osVersion": "{{properties.storageProfile.imageReference.sku}}",
		"tags": "{{tags}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Compute/virtualMachines/machine-name",
	"identity": null,
	"kind": "",
	"location": "northeurope",
	"managedBy": "",
	"name": "machine-name",
	"plan": null,
	"properties": {
		"extended": {
			"instanceView": {
				"computerName": "machine-name",
				"hyperVGeneration": "V1",
				"osName": "ubuntu",
				"osVersion": "20.04",
				"powerState": {
					"code": "PowerState/running",
					"displayStatus": "VM running",
					"level": "Info"
				}
			}
		},
		"hardwareProfile": {
			"vmSize": "Standard_D4s_v3"
		},
		"networkProfile": {
			"networkInterfaces": [
				{
					"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/networkInterfaces/machine-nameVMNic"
				}
			]
		},
		"osProfile": {
			"adminUsername": "admin",
			"allowExtensionOperations": true,
			"computerName": "machine-name",
			"linuxConfiguration": {
				"disablePasswordAuthentication": true,
				"patchSettings": {
					"assessmentMode": "ImageDefault",
					"patchMode": "ImageDefault"
				},
				"provisionVMAgent": true,
				"ssh": {}
			},
			"requireGuestProvisionSignal": true,
			"secrets": []
		},
		"provisioningState": "Succeeded",
		"storageProfile": {
			"dataDisks": [],
			"imageReference": {
				"exactVersion": "20.04.202505200",
				"offer": "0001-com-ubuntu-server-focal",
				"publisher": "Canonical",
				"sku": "20_04-lts",
				"version": "latest"
			},
			"osDisk": {
				"caching": "ReadWrite",
				"createOption": "FromImage",
				"deleteOption": "Detach",
				"diskSizeGB": 30,
				"managedDisk": {
					"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Compute/disks/machine-disk-name",
					"storageAccountType": "Premium_LRS"
				},
				"name": "machine-disk-name",
				"osType": "Linux"
			}
		},
		"timeCreated": "1970-01-01T00:00:00.00Z",
		"vmId": "898989-8989-8989-8989-898989898989"
	},
	"resourceGroup": "group-name",
	"sku": null,
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.compute/virtualmachines",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/get?view=rest-compute-2025-04-01&tabs=HTTP#virtualmachine
