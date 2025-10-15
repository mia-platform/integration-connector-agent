// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package main

import (
	"log"

	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
)

func main() {
	l, err := rpcprocessor.NewLogger("trace")
	if err != nil {
		log.Fatal(err)
	}

	processor := &CustomProcessor{
		logger: l,
	}
	rpcprocessor.Serve(&rpcprocessor.Config{
		Processor: processor,
		Logger:    l,
	})
}
