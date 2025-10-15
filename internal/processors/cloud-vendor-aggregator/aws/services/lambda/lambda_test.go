// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package lambda

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
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
		mockLambdaClient *awsclient.AWSMock
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
			mockLambdaClient: &awsclient.AWSMock{
				GetFunctionResult: &awsclient.Function{
					Name: "test-function",
					ARN:  "arn:aws:lambda:us-west-2:123456789012:function:test-function",
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
				client = &awsclient.AWSMock{}
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
