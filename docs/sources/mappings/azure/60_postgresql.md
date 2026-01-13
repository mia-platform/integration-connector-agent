# Microsoft Azure Database for PostgreSQL

For importing an Azure Database for PostgreSQL in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"state": "{{properties.state}}",
		"tags": "{{tags}}",
		"version": "{{properties.version}}",
		"minorVersion": "{{properties.minorVersion}}",
		"fullyQualifiedDomainName": "{{properties.fullyQualifiedDomainName}}",
		"publicNetworkAccess": "{{properties.network.publicNetworkAccess}}"	
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.DBforPostgreSQL/flexibleServers/server-name",
	"identity": null,
	"kind": "",
	"location": "northeurope",
	"managedBy": "",
	"name": "server-name",
	"plan": null,
	"properties": {
		"administratorLogin": "postgresql_admin",
		"authConfig": {
			"activeDirectoryAuth": "Disabled",
			"passwordAuth": "Enabled"
		},
		"availabilityZone": "2",
		"backup": {
			"backupRetentionDays": 7,
			"earliestRestoreDate": "1970-01-01T00:00:00.0000000Z",
			"geoRedundantBackup": "Disabled"
		},
		"dataEncryption": {
			"type": "SystemManaged"
		},
		"fullyQualifiedDomainName": "server-name.postgres.database.azure.com",
		"highAvailability": {
			"mode": "Disabled",
			"state": "NotEnabled"
		},
		"maintenanceWindow": {
			"customWindow": "Disabled",
			"dayOfWeek": 0,
			"startHour": 0,
			"startMinute": 0
		},
		"minorVersion": "10",
		"network": {
			"delegatedSubnetResourceId": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/virtualNetworks/azure-operationslab-vnet/subnets/default",
			"privateDnsZoneArmResourceId": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/privateDnsZones/server-name.private.postgres.database.azure.com",
			"publicNetworkAccess": "Disabled"
		},
		"replica": {
			"capacity": 5,
			"role": "Primary"
		},
		"replicaCapacity": 5,
		"replicationRole": "Primary",
		"state": "Ready",
		"storage": {
			"autoGrow": "Disabled",
			"iops": 120,
			"storageSizeGB": 32,
			"tier": "P4",
			"type": ""
		},
		"version": "16"
	},
	"resourceGroup": "group-name",
	"sku": {
		"name": "Standard_B1ms",
		"tier": "Burstable"
	},
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "71a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.dbforpostgresql/flexibleservers",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/postgresql/servers/get?view=rest-postgresql-2025-08-01&tabs=HTTP#server
