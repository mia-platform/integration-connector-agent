// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package config

// EnvironmentVariables struct with the mapping of desired
// environment variables.
type EnvironmentVariables struct {
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort             string `env:"HTTP_PORT" envDefault:"8080"`
	HTTPAddress          string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0"`
	ServicePrefix        string `env:"SERVICE_PREFIX"`
	DelayShutdownSeconds int    `env:"DELAY_SHUTDOWN_SECONDS" envDefault:"10"`

	ConfigurationPath string `env:"CONFIGURATION_PATH,required"`
}
