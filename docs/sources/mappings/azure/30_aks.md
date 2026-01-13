# Microsoft Azure Kubernetes Service (AKS)

For importing an Azure Kubernetes Service in the Catalog you can use this mapping configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"provisioningState": "{{properties.provisioningState}}",
		"tags": "{{tags}}",
		"currentKubernetesVersion": "{{properties.currentKubernetesVersion}}",
		"supportPlan": "{{properties.supportPlan}}",
		"publicNetworkAccess": "{{properties.publicNetworkAccess}}",
		"disableLocalAccounts": "{{properties.disableLocalAccounts}}",
		"dnsPrefix": "{{propertie.dnsPrefix}}",
		"fqdn": "{{properties.fqdn}}",
		"enableRBAC": "{{properties.enableRBAC}}",
		"skuTier": "{{sku.tier}}",
		"podCidr": "{{properties.networkProfile.podCidr}}",
		"serviceCidr": "{{properties.networkProfile.serviceCidr}}",
		"networkPlugin": "{{properties.networkProfile.networkPlugin}}",
		"outboundType": "{{properties.networkProfile.outboundType}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.ContainerService/managedClusters/cluster-name",
	"identity": {
		"principalId": "a042ad31-01c6-4b68-a65b-04354426754f",
		"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
		"type": "SystemAssigned"
	},
	"kind": "",
	"location": "northeurope",
	"managedBy": "",
	"name": "cluster-name",
	"plan": null,
	"properties": {
		"agentPoolProfiles": [
			{
				"count": 1,
				"currentOrchestratorVersion": "1.32.7",
				"enableAutoScaling": true,
				"enableEncryptionAtHost": false,
				"enableFIPS": false,
				"enableNodePublicIP": false,
				"enableUltraSSD": false,
				"kubeletDiskType": "OS",
				"maxCount": 1,
				"maxPods": 250,
				"minCount": 1,
				"mode": "System",
				"name": "default",
				"nodeImageVersion": "AKSUbuntu-2204containerd-202510.03.0",
				"orchestratorVersion": "1.32.7",
				"osDiskSizeGB": 128,
				"osDiskType": "Managed",
				"osSKU": "Ubuntu",
				"osType": "Linux",
				"powerState": {
					"code": "Running"
				},
				"provisioningState": "Succeeded",
				"scaleDownMode": "Delete",
				"securityProfile": {
					"enableSecureBoot": false,
					"enableVTPM": false
				},
				"type": "VirtualMachineScaleSets",
				"upgradeSettings": {
					"drainTimeoutInMinutes": 30,
					"maxSurge": "10%"
				},
				"vmSize": "Standard_A2_v2",
				"vnetSubnetID": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/virtualNetworks/network-name/subnets/subnet-name"
			},
			{
				"count": 2,
				"currentOrchestratorVersion": "1.32.7",
				"enableAutoScaling": true,
				"enableEncryptionAtHost": false,
				"enableFIPS": false,
				"enableNodePublicIP": false,
				"enableUltraSSD": false,
				"kubeletDiskType": "OS",
				"maxCount": 2,
				"maxPods": 250,
				"minCount": 1,
				"mode": "User",
				"name": "operations",
				"nodeImageVersion": "AKSUbuntu-2204gen2containerd-202510.03.0",
				"nodeLabels": {},
				"nodeTaints": [],
				"orchestratorVersion": "1.32.7",
				"osDiskSizeGB": 128,
				"osDiskType": "Managed",
				"osSKU": "Ubuntu",
				"osType": "Linux",
				"powerState": {
					"code": "Running"
				},
				"provisioningState": "Succeeded",
				"scaleDownMode": "Delete",
				"securityProfile": {
					"enableSecureBoot": false,
					"enableVTPM": false
				},
				"type": "VirtualMachineScaleSets",
				"upgradeSettings": {
					"drainTimeoutInMinutes": 30,
					"maxSurge": "10%"
				},
				"vmSize": "Standard_DS2_v2",
				"vnetSubnetID": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/virtualNetworks/network-name/subnets/subnet-name"
			}
		],
		"apiServerAccessProfile": {
			"authorizedIPRanges": [
				"0.0.0.0/0"
			]
		},
		"autoScalerProfile": {
			"balance-similar-node-groups": "false",
			"daemonset-eviction-for-empty-nodes": false,
			"daemonset-eviction-for-occupied-nodes": true,
			"expander": "random",
			"ignore-daemonsets-utilization": false,
			"max-empty-bulk-delete": "10",
			"max-graceful-termination-sec": "600",
			"max-node-provision-time": "15m",
			"max-total-unready-percentage": "45",
			"new-pod-scale-up-delay": "0s",
			"ok-total-unready-count": "3",
			"scale-down-delay-after-add": "10m",
			"scale-down-delay-after-delete": "10s",
			"scale-down-delay-after-failure": "3m",
			"scale-down-unneeded-time": "10m",
			"scale-down-unready-time": "20m",
			"scale-down-utilization-threshold": "0.5",
			"scan-interval": "10s",
			"skip-nodes-with-local-storage": "false",
			"skip-nodes-with-system-pods": "false"
		},
		"autoUpgradeProfile": {
			"nodeOSUpgradeChannel": "NodeImage",
			"upgradeChannel": "none"
		},
		"azureMonitorProfile": {
			"metrics": {
				"enabled": false,
				"kubeStateMetrics": {}
			}
		},
		"azurePortalFQDN": "cluster-prefix-00000.portal.hcp.northeurope.azmk8s.io",
		"bootstrapProfile": {
			"artifactSource": "Direct"
		},
		"currentKubernetesVersion": "1.32.7",
		"disableLocalAccounts": false,
		"dnsPrefix": "cluster-prefix",
		"enableRBAC": true,
		"fqdn": "cluster-prefix-00000.hcp.northeurope.azmk8s.io",
		"identityProfile": {
			"kubeletidentity": {
				"clientId": "12121212-1212-1212-1212-121212121212",
				"objectId": "24242424-2424-2424-2424-242424242424",
				"resourceId": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourcegroups/node-group-name/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cluster-name-agentpool"
			}
		},
		"kubernetesVersion": "1.32.7",
		"maxAgentPools": 100,
		"metricsProfile": {
			"costAnalysis": {
				"enabled": false
			}
		},
		"networkProfile": {
			"dnsServiceIP": "10.0.0.10",
			"ipFamilies": [
				"IPv4"
			],
			"loadBalancerProfile": {
				"backendPoolType": "nodeIPConfiguration",
				"effectiveOutboundIPs": [
					{
						"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/publicIPAddresses/ip-address-name"
					}
				],
				"idleTimeoutInMinutes": 30,
				"outboundIPs": {
					"publicIPs": [
						{
							"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/group-name/providers/Microsoft.Network/publicIPAddresses/ip-address-name"
						}
					]
				}
			},
			"loadBalancerSku": "standard",
			"networkPlugin": "none",
			"networkPolicy": "none",
			"outboundType": "loadBalancer",
			"serviceCidr": "10.0.0.0/16",
			"serviceCidrs": [
				"10.0.0.0/16"
			]
		},
		"nodeResourceGroup": "node-group-name",
		"oidcIssuerProfile": {
			"enabled": true,
			"issuerURL": "https://northeurope.oic.prod-aks.azure.com/1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a/2b2b2b2b-2b2b-2b2b-2b2b-2b2b2b2b2b2b/"
		},
		"powerState": {
			"code": "Running"
		},
		"provisioningState": "Succeeded",
		"resourceUID": "000000000000000000000000",
		"securityProfile": {
			"imageCleaner": {
				"enabled": false,
				"intervalHours": 48
			}
		},
		"servicePrincipalProfile": {
			"clientId": "msi"
		},
		"storageProfile": {
			"diskCSIDriver": {
				"enabled": true
			},
			"fileCSIDriver": {
				"enabled": true
			},
			"snapshotController": {
				"enabled": true
			}
		},
		"supportPlan": "KubernetesOfficial",
		"upgradeSettings": {
			"overrideSettings": {
				"forceUpgrade": false
			}
		},
		"windowsProfile": {
			"adminUsername": "azureuser",
			"enableCSIProxy": true
		},
		"workloadAutoScalerProfile": {}
	},
	"resourceGroup": "group-name",
	"sku": {
		"name": "Base",
		"tier": "Free"
	},
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": null,
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.containerservice/managedclusters",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/aks/managed-clusters/get?view=rest-aks-2025-08-01&tabs=HTTP#managedcluster
