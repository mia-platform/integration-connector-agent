// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package consoleclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type consoleClient[T Resource] struct {
	url string
	tm  TokenManager
}

func New[T Resource](url string, tm TokenManager) CatalogClient[T] {
	return &consoleClient[T]{url: url, tm: tm}
}

func (c *consoleClient[T]) fireRequest(_ context.Context, verb, targetURL string, requestBody any) (*http.Response, error) {
	bodyReader, err := prepareBody(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(verb, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarketplaceRequestCreation, err)
	}

	if err := c.tm.SetAuthHeader(req); err != nil {
		return nil, err
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarketplaceRequestExecution, err)
	}

	return resp, nil
}

func prepareBody(requestBody any) (io.Reader, error) {
	if requestBody == nil {
		return nil, nil
	}
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarketplaceRequestBodyParse, err)
	}

	return bytes.NewReader(reqBodyBytes), nil
}
