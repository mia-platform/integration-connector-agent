// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package server

import "errors"

var (
	errSetupSource       = errors.New("error setting up source")
	errSetupWriter       = errors.New("error setting up writer")
	errUnsupportedWriter = errors.New("unsupported writer type")
)
