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

package utils

import (
	"fmt"
	"runtime"
)

var (
	// Version  is dynamically set at build time
	Version = "DEV"
	// BuildDate is dynamically set at build time
	BuildDate = "" // YYYY-MM-DD
)

func ServiceVersionInformation() string {
	version := Version

	if BuildDate != "" {
		version = fmt.Sprintf("%s (%s)", version, BuildDate)
	}

	return fmt.Sprintf("%s, Go Version: %s", version, runtime.Version())
}
