// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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
