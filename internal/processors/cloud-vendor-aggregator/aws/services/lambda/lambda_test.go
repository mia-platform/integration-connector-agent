// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambda

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/lambda"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestGetData(t *testing.T) {
	testCases := []struct {
		name             string
		event            *awssqsevents.CloudTrailEvent
		expectedError    error
		expectedAsset    *commons.Asset
		mockLambdaClient *mockLambdaClient
	}{
		{
			name: "error if event is missing the functionName in requestParameters and responseElements",
			event: &awssqsevents.CloudTrailEvent{
				Source: "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource:       "lambda.amazonaws.com",
					AWSRegion:         "us-west-2",
					RequestParameters: map[string]interface{}{},
					ResponseElements:  map[string]interface{}{},
				},
			},
			expectedError: commons.ErrInvalidEvent,
		},
		{
			name: "returns asset with functionName from responseElements if missing from requestParameters",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "lambda.amazonaws.com",
					AWSRegion:   "us-west-2",
					ResponseElements: map[string]interface{}{
						"functionName": "test-function",
					},
					RequestParameters: map[string]interface{}{},
				},
			},
			expectedAsset: &commons.Asset{
				Name:          "test-function",
				Type:          "lambda.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags:          commons.Tags{},
			},
		},
		{
			name: "returns asset with functionName from requestParameters",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "lambda.amazonaws.com",
					AWSRegion:   "us-west-2",
					RequestParameters: map[string]interface{}{
						"functionName": "test-function",
					},
				},
			},
			expectedAsset: &commons.Asset{
				Name:          "test-function",
				Type:          "lambda.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags:          commons.Tags{},
			},
		},
		{
			name: "returns asset with functionName from requestParameters if value in responseElements is not a string",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "lambda.amazonaws.com",
					AWSRegion:   "us-west-2",
					ResponseElements: map[string]interface{}{
						"functionName": 123, // not a string
					},
					RequestParameters: map[string]interface{}{
						"functionName": "test-function",
					},
				},
			},
			expectedAsset: &commons.Asset{
				Name:          "test-function",
				Type:          "lambda.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags:          commons.Tags{},
			},
		},
		{
			name: "returns function name from arn resource on tag events",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource:      "lambda.amazonaws.com",
					AWSRegion:        "us-west-2",
					ResponseElements: nil,
					EventName:        "TagResource20170331v2",
					RequestParameters: map[string]interface{}{
						"resource": "arn:aws:lambda:eu-north-1:accountid:function:test-function",
						"tags": map[string]string{
							"t1": "tv2",
						},
					},
				},
			},
			expectedAsset: &commons.Asset{
				Name:          "test-function",
				Type:          "lambda.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags:          commons.Tags{},
			},
		},
		{
			name: "returns tags from function if available",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.lambda",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "lambda.amazonaws.com",
					AWSRegion:   "us-west-2",
					RequestParameters: map[string]interface{}{
						"functionName": "test-function",
					},
				},
			},
			mockLambdaClient: &mockLambdaClient{
				function: &lambda.Function{
					FunctionName: "test-function",
					FunctionArn:  "arn:aws:lambda:us-west-2:123456789012:function:test-function",
					Tags: commons.Tags{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			expectedAsset: &commons.Asset{
				Name:          "test-function",
				Type:          "lambda.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags: commons.Tags{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, _ := test.NewNullLogger()

			client := tc.mockLambdaClient
			if client == nil {
				client = &mockLambdaClient{}
			}

			lambda := New(l, client)
			newData, err := lambda.GetData(t.Context(), tc.event)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)

				var asset commons.Asset
				require.NoError(t, json.Unmarshal(newData, &asset))
				require.Equal(t, tc.expectedAsset.Name, asset.Name)
				require.Equal(t, tc.expectedAsset.Type, asset.Type)
				require.Equal(t, tc.expectedAsset.Location, asset.Location)
				require.Equal(t, tc.expectedAsset.Provider, asset.Provider)
				require.Equal(t, tc.expectedAsset.Tags, asset.Tags)
				require.NotEmpty(t, asset.Timestamp)
			}
		})
	}
}

type mockLambdaClient struct {
	function    *lambda.Function
	functionErr error
}

func (m *mockLambdaClient) GetFunction(_ context.Context, _ string) (*lambda.Function, error) {
	return m.function, m.functionErr
}
