// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package azure

import (
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type ActivityLogEventData struct {
	Records []*ActivityLogEventRecord `json:"records,omitempty"`
}

type ActivityLogEventRecord struct {
	RoleLocation    string         `json:"RoleLocation,omitempty"`   //nolint:tagliatelle
	Stamp           string         `json:"Stamp,omitempty"`          //nolint:tagliatelle
	ReleaseVersion  string         `json:"ReleaseVersion,omitempty"` //nolint:tagliatelle
	Time            string         `json:"time,omitempty"`
	ResourceID      string         `json:"resourceId,omitempty"`
	OperationName   string         `json:"operationName,omitempty"`
	Category        string         `json:"category,omitempty"`
	ResultType      string         `json:"resultType,omitempty"`
	ResultSignature string         `json:"resultSignature,omitempty"`
	DurationMs      string         `json:"durationMs,omitempty"`
	CallerIPAddress string         `json:"callerIpAddress,omitempty"`
	CorrelationID   string         `json:"correlationId,omitempty"`
	Level           string         `json:"level,omitempty"`
	Properties      map[string]any `json:"properties,omitempty"`
}

func (r *ActivityLogEventRecord) entityOperationType() entities.Operation {
	if strings.HasSuffix(strings.ToLower(r.OperationName), "delete") ||
		strings.HasSuffix(strings.ToLower(r.OperationName), "delete/action") {
		return entities.Delete
	}

	return entities.Write
}
