// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/entities"
)

//go:generate ${TOOLS_BIN}/stringer -type=EventType -trimprefix=EventType
type EventType int

const (
	EventTypeRecordFromEventHub EventType = iota
	EventTypeFromLiveLoad
)

func EventFromRecord(record *ActivityLogEventRecord) *entities.Event {
	rawRecord, err := json.Marshal(record)
	if err != nil {
		return nil
	}

	return &entities.Event{
		PrimaryKeys:   primaryKeys(record.ResourceID),
		OperationType: record.entityOperationType(),
		Type:          EventTypeRecordFromEventHub.String(),
		OriginalRaw:   rawRecord,
	}
}
