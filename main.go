// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
//
// This file is part of integration-connector-agent.
//
// integration-connector-agent is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// integration-connector-agent is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with integration-connector-agent. If not, see <https://www.gnu.org/licenses/>.
//
// Alternatively, this file may be used under the terms of a commercial license
// available from Mia-Platform. For inquiries, contact licensing@mia-platform.eu.

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
