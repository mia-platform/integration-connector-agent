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

package awssqsevents

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type CloudTrailEventDetail struct {
	EventVersion string `json:"eventVersion"`
	UserIdentity struct {
		Type           string `json:"type"`
		PrincipalID    string `json:"principalId"`
		ARN            string `json:"arn"`
		AccountID      string `json:"accountId"`
		AccessKeyID    string `json:"accessKeyId"`
		SessionContext struct {
			Attributes    map[string]any `json:"attributes"`
			SessionIssuer map[string]any `json:"sessionIssuer"`
		}
	}
	EventTime          string         `json:"eventTime"`
	EventSource        string         `json:"eventSource"`
	EventName          string         `json:"eventName"`
	AWSRegion          string         `json:"awsRegion"`
	SourceIPAddress    string         `json:"sourceIPAddress"` //nolint:tagliatelle
	UserAgent          string         `json:"userAgent"`
	ErrorCode          string         `json:"errorCode"`
	ErrorMessage       string         `json:"errorMessage"`
	RequestParameters  map[string]any `json:"requestParameters"`
	ResponseElements   map[string]any `json:"responseElements"`
	RequestID          string         `json:"requestID"` //nolint:tagliatelle
	EventID            string         `json:"eventID"`   //nolint:tagliatelle
	ReadOnly           bool           `json:"readOnly"`
	EventType          string         `json:"eventType"`
	ManagementEvent    bool           `json:"managementEvent"`
	RecipientAccountID string         `json:"recipientAccountId"`
	EventCategory      string         `json:"eventCategory"`
	TLSDetails         struct {
		CipherSuite              string `json:"cipherSuite"`
		TLSVersion               string `json:"tlsVersion"`
		ClientProvidedHostHeader string `json:"clientProvidedHostHeader"`
	} `json:"tlsDetails"`
	SessionCredentialFromConsole string `json:"sessionCredentialFromConsole"`
}

type CloudTrailEvent struct {
	Version    string                `json:"version"`
	ID         string                `json:"id"`
	DetailType string                `json:"detail-type"` //nolint:tagliatelle
	Source     string                `json:"source"`
	Account    string                `json:"account"`
	Time       string                `json:"time"`
	Region     string                `json:"region"`
	Detail     CloudTrailEventDetail `json:"detail"`
}

var eventMap = map[string]struct {
	resourceNameField string
	// resourceNameFromResponseEvents contains event names where the resource name
	// is found in the response elements instead of request parameters.
	resourceNameFromResponseEvents []string

	resourceNameFromResourceArnEvents []string
}{
	"aws.s3": {
		resourceNameField: "bucketName",
	},
	"aws.lambda": {
		resourceNameField:                 "functionName",
		resourceNameFromResponseEvents:    []string{"UpdateFunctionCode20150331v2"},
		resourceNameFromResourceArnEvents: []string{"TagResource20170331v2"},
	},
}

func (e CloudTrailEvent) ResourceName() (string, error) {
	eventMappedData, ok := eventMap[e.Source]
	if !ok {
		return "", fmt.Errorf("unsupported event source: %s", e.Source)
	}

	if slices.Contains(eventMappedData.resourceNameFromResourceArnEvents, e.Detail.EventName) {
		resource, exists := e.Detail.RequestParameters["resource"]
		if !exists {
			return "", fmt.Errorf("resource field not found in event detail")
		}
		resourceArn, ok := resource.(string)
		if !ok {
			return "", fmt.Errorf("resource field is not a string")
		}

		if !arn.IsARN(resourceArn) {
			return "", fmt.Errorf("resource field is not a valid ARN: %s", resourceArn)
		}
		parsedARN, err := arn.Parse(resourceArn)
		if err != nil {
			return "", fmt.Errorf("error parsing resource ARN: %w", err)
		}

		tokens := strings.Split(parsedARN.Resource, ":")
		return tokens[1], nil
	}

	resourceNameField := eventMappedData.resourceNameField

	if e.Detail.ResponseElements != nil {
		if name, ok := e.Detail.ResponseElements[resourceNameField]; ok {
			nameStr, ok := name.(string)
			if ok {
				return nameStr, nil
			}
		}
	}

	if e.Detail.RequestParameters != nil {
		name := e.Detail.RequestParameters[resourceNameField]
		nameStr, ok := name.(string)
		if ok {
			return nameStr, nil
		}
	}

	return "", fmt.Errorf("resource name field %s not found in event detail", resourceNameField)
}

func (e CloudTrailEvent) EventSource() string {
	return e.Detail.EventSource
}

func (e CloudTrailEvent) Operation() (entities.Operation, error) {
	eventName := e.Detail.EventName
	switch {
	case strings.HasPrefix(eventName, "Delete"):
		return entities.Delete, nil

	case strings.HasPrefix(eventName, "Create"):
		return entities.Write, nil

	case strings.HasPrefix(eventName, "Update"):
		return entities.Write, nil

	case strings.HasPrefix(eventName, "Publish"):
		return entities.Write, nil

	case strings.HasPrefix(eventName, "Put"):
		return entities.Write, nil

	case strings.HasPrefix(eventName, "Tag"):
		return entities.Write, nil

	default:
		return entities.Write, fmt.Errorf("unsupported event name: %s", eventName)
	}
}

func (e CloudTrailEvent) EventType() string {
	return RealtimeSyncEventType
}

func (e CloudTrailEvent) GetRegion() string {
	return e.Detail.AWSRegion
}

func (e CloudTrailEvent) AccountID() string {
	return "account/" + e.Account
}

// -----

type CloudTrailImportEvent struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Region  string `json:"region"`
	Account string `json:"account"`
}

func (e CloudTrailImportEvent) ResourceName() (string, error) {
	return e.Name, nil
}

func (e CloudTrailImportEvent) EventSource() string {
	return e.Source
}

func (e CloudTrailImportEvent) Operation() (entities.Operation, error) {
	return entities.Write, nil
}

func (e CloudTrailImportEvent) EventType() string {
	return ImportEventType
}

func (e CloudTrailImportEvent) GetRegion() string {
	return e.Region
}

func (e CloudTrailImportEvent) AccountID() string {
	return "account/" + e.Account
}
