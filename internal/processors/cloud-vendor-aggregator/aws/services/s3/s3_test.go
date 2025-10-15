// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package s3

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	aws "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestGetData(t *testing.T) {
	testCases := []struct {
		name          string
		event         *awssqsevents.CloudTrailEvent
		expectedError error
		expectedAsset *commons.Asset
		mockS3Client  *aws.AWSMock
	}{
		{
			name: "error if event is missing the bucketName in request parameters",
			event: &awssqsevents.CloudTrailEvent{
				Source: "aws.s3",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "s3.amazonaws.com",
					AWSRegion:   "us-west-2",
				},
			},
			expectedError: commons.ErrInvalidEvent,
		},
		{
			name: "returns tags retrieved from S3",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.s3",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "s3.amazonaws.com",
					AWSRegion:   "us-west-2",
					RequestParameters: map[string]interface{}{
						"bucketName": "test-bucket",
					},
				},
			},
			mockS3Client: &aws.AWSMock{
				GetBucketTagsResult: commons.Tags{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expectedAsset: &commons.Asset{
				Relationships: []string{"account/123456789012"},
				Name:          "test-bucket",
				Type:          "s3.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Tags: commons.Tags{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "returns empty tags if S3 client fails",
			event: &awssqsevents.CloudTrailEvent{
				Account: "123456789012",
				Source:  "aws.s3",
				Detail: awssqsevents.CloudTrailEventDetail{
					EventSource: "s3.amazonaws.com",
					AWSRegion:   "us-west-2",
					RequestParameters: map[string]interface{}{
						"bucketName": "test-bucket",
					},
				},
			},
			mockS3Client: &aws.AWSMock{
				GetBucketTagsError: errors.New("failed to get tags"),
			},
			expectedAsset: &commons.Asset{
				Name:          "test-bucket",
				Type:          "s3.amazonaws.com",
				Location:      "us-west-2",
				Provider:      commons.AWSAssetProvider,
				Relationships: []string{"account/123456789012"},
				Tags:          commons.Tags{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, _ := test.NewNullLogger()

			s3client := tc.mockS3Client
			if s3client == nil {
				s3client = &aws.AWSMock{}
			}

			s3 := New(l, s3client)
			newData, err := s3.GetData(t.Context(), tc.event)

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
