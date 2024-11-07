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
