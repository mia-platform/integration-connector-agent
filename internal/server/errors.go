// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package server

import "errors"

var (
	errSetupSource       = errors.New("error setting up source")
	errSetupWriter       = errors.New("error setting up writer")
	errUnsupportedWriter = errors.New("unsupported writer type")
)
