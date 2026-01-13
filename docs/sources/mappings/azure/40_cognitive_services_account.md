# Microsoft Azure Cognitive Services Account

For importing an Azure Cognitive Services Account in the Catalog you can use this mapping
configuration:

```json
{
	"type": "mapper",
	"outputEvent": {
		"id": "{{id}}",
		"name": "{{name}}",
		"location": "{{location}}",
		"provisioningState": "{{properties.provisioningState}}",
		"tags": "{{tags}}",
		"kind": "{{kind}}",
		"endpoint": "{{properties.endpoint}}"
	}
}
```

If you want to use a custom mapping or you want to add other values to the mapping, you can refer
this resource example of visit the [official documentation site]:

```json
{
	"extendedLocation": null,
	"id": "/subscriptions/0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f/resourceGroups/resource-group/providers/Microsoft.CognitiveServices/accounts/account-name",
	"identity": null,
	"kind": "OpenAI",
	"location": "francecentral",
	"managedBy": "",
	"name": "account-name",
	"plan": null,
	"properties": {
		"apiProperties": {},
		"callRateLimit": {
			"rules": [
				{
					"count": 100,
					"key": "openai.batches.list",
					"matchPatterns": [
						{
							"method": "GET",
							"path": "openai/batches"
						},
						{
							"method": "GET",
							"path": "openai/v1/batches"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 100000,
					"key": "openai.assistants.default",
					"matchPatterns": [
						{
							"method": "*",
							"path": "openai/assistants"
						},
						{
							"method": "*",
							"path": "openai/assistants/*"
						},
						{
							"method": "*",
							"path": "openai/threads"
						},
						{
							"method": "*",
							"path": "openai/threads/*"
						},
						{
							"method": "*",
							"path": "openai/vector_stores"
						},
						{
							"method": "*",
							"path": "openai/vector_stores/*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 100000,
					"key": "openai.responses.default",
					"matchPatterns": [
						{
							"method": "*",
							"path": "openai/responses"
						},
						{
							"method": "*",
							"path": "openai/responses/*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 120,
					"key": "openai.assistants.list",
					"matchPatterns": [
						{
							"method": "GET",
							"path": "openai/assistants"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 120,
					"key": "openai.threads.list",
					"matchPatterns": [
						{
							"method": "GET",
							"path": "openai/threads"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 120,
					"key": "openai.vectorstores.list",
					"matchPatterns": [
						{
							"method": "GET",
							"path": "openai/vector_stores"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 30,
					"key": "default",
					"matchPatterns": [
						{
							"method": "*",
							"path": "*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 30,
					"key": "openai",
					"matchPatterns": [
						{
							"method": "*",
							"path": "openai/*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 30,
					"key": "openai.batches.post",
					"matchPatterns": [
						{
							"method": "POST",
							"path": "openai/batches"
						},
						{
							"method": "POST",
							"path": "openai/v1/batches"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 30,
					"key": "openai.dalle.other",
					"matchPatterns": [
						{
							"method": "*",
							"path": "dalle/*"
						},
						{
							"method": "*",
							"path": "openai/operations/images/*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 30,
					"key": "openai.dalle.post",
					"matchPatterns": [
						{
							"method": "POST",
							"path": "dalle/*"
						},
						{
							"method": "POST",
							"path": "openai/images/*"
						}
					],
					"renewalPeriod": 1
				},
				{
					"count": 500,
					"key": "openai.batches.get",
					"matchPatterns": [
						{
							"method": "GET",
							"path": "openai/batches/*"
						},
						{
							"method": "GET",
							"path": "openai/v1/batches/*"
						}
					],
					"renewalPeriod": 60
				},
				{
					"count": 60,
					"key": "openai.vectorstores.post",
					"matchPatterns": [
						{
							"method": "POST",
							"path": "openai/vector_stores"
						},
						{
							"method": "POST",
							"path": "openai/vector_stores/*"
						}
					],
					"renewalPeriod": 1
				}
			]
		},
		"capabilities": [
			{
				"name": "CustomerManagedKey"
			},
			{
				"name": "EnableRLSForThrottling",
				"value": "false"
			},
			{
				"name": "MaxEvaluationRunDurationInHours",
				"value": "5"
			},
			{
				"name": "MaxFineTuneCount",
				"value": "500"
			},
			{
				"name": "MaxFineTuneJobDurationInHours",
				"value": "720"
			},
			{
				"name": "MaxRunningEvaluationCount",
				"value": "5"
			},
			{
				"name": "MaxRunningFineTuneCount",
				"value": "3"
			},
			{
				"name": "MaxRunningGlobalStandardFineTuneCount",
				"value": "3"
			},
			{
				"name": "MaxTrainingFileSize",
				"value": "512000000"
			},
			{
				"name": "MaxUserFileCount",
				"value": "100"
			},
			{
				"name": "MaxUserFileImportDurationInHours",
				"value": "1"
			},
			{
				"name": "RaiMonitor"
			},
			{
				"name": "TrustedServices",
				"value": "Microsoft.CognitiveServices,Microsoft.MachineLearningServices,Microsoft.Search,Microsoft.VideoIndexer"
			},
			{
				"name": "VirtualNetworks"
			},
			{
				"name": "quotaRefundsEnabled",
				"value": "false"
			}
		],
		"customSubDomainName": "account-name",
		"dateCreated": "1970-01-01T00:00:00.0000000Z",
		"endpoint": "https://account-name.openai.azure.com/",
		"endpoints": {
			"Azure OpenAI Legacy API - Latest moniker": "https://account-name.openai.azure.com/",
			"OpenAI Dall-E API": "https://account-name.openai.azure.com/",
			"OpenAI Language Model Instance API": "https://account-name.openai.azure.com/",
			"OpenAI Model Scaleset API": "https://account-name.openai.azure.com/",
			"OpenAI Moderations API": "https://account-name.openai.azure.com/",
			"OpenAI Realtime API": "https://account-name.openai.azure.com/",
			"OpenAI Sora API": "https://account-name.openai.azure.com/",
			"OpenAI Whisper API": "https://account-name.openai.azure.com/",
			"Token Service API": "https://account-name.openai.azure.com/"
		},
		"internalId": "00000000000000000000000000000000",
		"isMigrated": false,
		"networkAcls": {
			"defaultAction": "Allow",
			"ipRules": [],
			"virtualNetworkRules": []
		},
		"privateEndpointConnections": [],
		"provisioningState": "Succeeded",
		"publicNetworkAccess": "Enabled"
	},
	"resourceGroup": "rocket",
	"sku": {
		"name": "S0"
	},
	"subscriptionId": "0f0f0f0f-0f0f-0f0f-0f0f-0f0f0f0f0f0f",
	"tags": {},
	"tenantId": "1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a",
	"type": "microsoft.cognitiveservices/accounts",
	"zones": null
}
```

[official documentation site]: https://learn.microsoft.com/en-us/rest/api/aiservices/accountmanagement/accounts/get?view=rest-aiservices-accountmanagement-2024-10-01&tabs=HTTP#account
