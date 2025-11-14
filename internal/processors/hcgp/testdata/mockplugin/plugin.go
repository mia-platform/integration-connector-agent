// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package main

import (
	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
)

func main() {
	logger, _ := rpcprocessor.NewLogger("trace")

	processor := &MockProcessor{
		logger: logger,
	}
	rpcprocessor.Serve(&rpcprocessor.Config{
		Processor: processor,
		Logger:    logger,
	})
}
