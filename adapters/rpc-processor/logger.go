// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package rpcprocessor

import (
	"github.com/mia-platform/integration-connector-agent/internal/processors/hcgp"

	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

type Logger = hcgp.Logger

func NewLogger(level string) (Logger, error) {
	l, err := glogrus.InitHelper(glogrus.InitOptions{
		Level: level,
	})

	if err != nil {
		return nil, err
	}
	return l, nil
}
