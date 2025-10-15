// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package config

import (
	"bytes"
	"os"
)

func readFile(path string) ([]byte, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(configFile); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
