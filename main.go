// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/server"

	"github.com/caarlos0/env/v11"
)

func main() {
	envVars, err := env.ParseAs[config.EnvironmentVariables]()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config, err := config.LoadServiceConfiguration(envVars.ConfigurationPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sysChan := make(chan os.Signal, 1)
	signal.Notify(sysChan, syscall.SIGTERM)
	exitCode := 0

	if err := server.New(context.Background(), envVars, config, sysChan); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	}

	close(sysChan)
	os.Exit(exitCode)
}
