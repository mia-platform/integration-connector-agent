// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package utils

import "errors"

var (
	ErrValidationError = errors.New("validation error")
)

type HTTPError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func ValidationError(message string) *HTTPError {
	return &HTTPError{
		Error:   "Validation Error",
		Message: message,
	}
}

func InternalServerError(message string) *HTTPError {
	return &HTTPError{
		Error:   "Internal Server Error",
		Message: message,
	}
}
