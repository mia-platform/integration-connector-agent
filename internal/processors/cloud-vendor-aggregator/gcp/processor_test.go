// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpaggregator

import (
	"encoding/json"
	"os"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestProcessWithStandardAssetInventoryEvent(t *testing.T) {
	l, _ := test.NewNullLogger()

	testCases := []struct {
		name          string
		eventFilePath string
		expectedAsset assetpb.Asset
	}{
		{
			name:          "Create bucket event",
			eventFilePath: "testdata/bucket-create.json",
			expectedAsset: assetpb.Asset{
				Name:      "//storage.googleapis.com/test-bucket",
				AssetType: "storage.googleapis.com/Bucket",
				Resource: &assetpb.Resource{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"id":   structpb.NewStringValue("test-bucket"),
							"name": structpb.NewStringValue("test-bucket"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"label1": structpb.NewStringValue("value1"),
									"label2": structpb.NewStringValue("value2"),
								},
							}),
						},
					},
				},
			},
		},
		{
			name:          "Create network event",
			eventFilePath: "testdata/network-create.json",
			expectedAsset: assetpb.Asset{
				Name:      "//compute.googleapis.com/test-network",
				AssetType: "compute.googleapis.com/Network",
				Resource: &assetpb.Resource{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"id":   structpb.NewStringValue("test-network"),
							"name": structpb.NewStringValue("test-network"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"label1": structpb.NewStringValue("value1"),
									"label2": structpb.NewStringValue("value2"),
								},
							}),
						},
					},
				},
			},
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			p := &GCPCloudVendorAggregator{
				logger:  l,
				options: nil,
			}

			d, err := os.ReadFile(tc.eventFilePath)
			require.NoError(t, err)

			builder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryEvent]()
			event, err := builder.GetPipelineEvent(t.Context(), d)
			require.NoError(t, err)

			result, err := p.Process(event)
			require.NoError(t, err)
			require.NotNil(t, result)

			var asset assetpb.Asset
			require.NoError(t, protojson.Unmarshal(result.Data(), &asset))

			assetData := asset.GetResource().GetData()
			idField := assetData.GetFields()["id"]
			nameField := assetData.GetFields()["name"]
			labelsField := assetData.GetFields()["labels"]

			expectedAssetData := tc.expectedAsset.GetResource().GetData()
			expectedIDField := expectedAssetData.GetFields()["id"]
			expectedNameField := expectedAssetData.GetFields()["name"]
			expectedLabelsField := expectedAssetData.GetFields()["labels"]

			require.Equal(t, tc.expectedAsset.GetName(), asset.GetName())
			require.Equal(t, tc.expectedAsset.GetAssetType(), asset.GetAssetType())
			require.Equal(t, expectedIDField.GetStringValue(), idField.GetStringValue())
			require.Equal(t, expectedNameField.GetStringValue(), nameField.GetStringValue())
			require.Equal(t, expectedLabelsField.GetStructValue().AsMap(), labelsField.GetStructValue().AsMap())
		})
	}
}

func TestProcessWithImportAssetInventoryEvent(t *testing.T) {
	l, _ := test.NewNullLogger()

	testCases := []struct {
		name          string
		importEvent   gcppubsubevents.InventoryImportEvent
		expectedAsset assetpb.Asset
	}{
		{
			name: "import bucket event",
			importEvent: gcppubsubevents.InventoryImportEvent{
				AssetName: "//storage.googleapis.com/test-bucket",
				Type:      "storage.googleapis.com/Bucket",
				Data: &assetpb.Asset{
					Name:      "//storage.googleapis.com/test-bucket",
					AssetType: "storage.googleapis.com/Bucket",
					Resource: &assetpb.Resource{
						Data: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"id":   structpb.NewStringValue("test-bucket"),
								"name": structpb.NewStringValue("test-bucket"),
								"labels": structpb.NewStructValue(&structpb.Struct{
									Fields: map[string]*structpb.Value{
										"label1": structpb.NewStringValue("value1"),
										"label2": structpb.NewStringValue("value2"),
									},
								}),
							},
						},
					},
				},
			},
			expectedAsset: assetpb.Asset{
				Name:      "//storage.googleapis.com/test-bucket",
				AssetType: "storage.googleapis.com/Bucket",
				Resource: &assetpb.Resource{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"id":   structpb.NewStringValue("test-bucket"),
							"name": structpb.NewStringValue("test-bucket"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"label1": structpb.NewStringValue("value1"),
									"label2": structpb.NewStringValue("value2"),
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "import network event",
			importEvent: gcppubsubevents.InventoryImportEvent{
				AssetName: "//compute.googleapis.com/test-network",
				Type:      "compute.googleapis.com/Network",
				Data: &assetpb.Asset{
					Name:      "//compute.googleapis.com/test-network",
					AssetType: "compute.googleapis.com/Network",
					Resource: &assetpb.Resource{
						Data: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"id":   structpb.NewStringValue("test-network"),
								"name": structpb.NewStringValue("test-network"),
								"labels": structpb.NewStructValue(&structpb.Struct{
									Fields: map[string]*structpb.Value{
										"label1": structpb.NewStringValue("value1"),
										"label2": structpb.NewStringValue("value2"),
									},
								}),
							},
						},
					},
				},
			},
			expectedAsset: assetpb.Asset{
				Name:      "//compute.googleapis.com/test-network",
				AssetType: "compute.googleapis.com/Network",
				Resource: &assetpb.Resource{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"id":   structpb.NewStringValue("test-network"),
							"name": structpb.NewStringValue("test-network"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"label1": structpb.NewStringValue("value1"),
									"label2": structpb.NewStringValue("value2"),
								},
							}),
						},
					},
				},
			},
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			p := &GCPCloudVendorAggregator{
				logger:  l,
				options: nil,
			}

			data, err := json.Marshal(tc.importEvent)
			require.NoError(t, err)

			builder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryImportEvent]()
			event, err := builder.GetPipelineEvent(t.Context(), data)
			require.NoError(t, err)

			result, err := p.Process(event)
			require.NoError(t, err)
			require.NotNil(t, result)

			var asset assetpb.Asset
			require.NoError(t, json.Unmarshal(result.Data(), &asset))

			assetData := asset.GetResource().GetData()
			idField := assetData.GetFields()["id"]
			nameField := assetData.GetFields()["name"]
			labelsField := assetData.GetFields()["labels"]

			expectedAssetData := tc.expectedAsset.GetResource().GetData()
			expectedIDField := expectedAssetData.GetFields()["id"]
			expectedNameField := expectedAssetData.GetFields()["name"]
			expectedLabelsField := expectedAssetData.GetFields()["labels"]

			require.Equal(t, tc.expectedAsset.GetName(), asset.GetName())
			require.Equal(t, tc.expectedAsset.GetAssetType(), asset.GetAssetType())
			require.Equal(t, expectedIDField.GetStringValue(), idField.GetStringValue())
			require.Equal(t, expectedNameField.GetStringValue(), nameField.GetStringValue())
			require.Equal(t, expectedLabelsField.GetStructValue().AsMap(), labelsField.GetStructValue().AsMap())
		})
	}
}
