// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package consoleclient

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
)

type MockExpectation struct {
	path       string
	verb       string
	tenantID   string
	headers    map[string]string
	bodyString string
}

type MockResponse struct {
	statusCode int
	body       any
	times      int
}

func runMocha(t *testing.T, mockAddr string) *mocha.Mocha {
	t.Helper()

	options := mocha.Configure().Addr(mockAddr)
	if testing.Verbose() {
		options = options.LogVerbosity(mocha.LogVerbose)
	}

	m := mocha.New(t, options.Build())
	m.CloseOnCleanup(t)
	m.Start()

	return m
}

func registerAPI(t *testing.T, m *mocha.Mocha, request MockExpectation, responses ...MockResponse) *mocha.Mocha {
	t.Helper()

	pathMatcher := expect.URLPath(request.path)
	var mock *mocha.MockBuilder
	switch request.verb {
	case http.MethodPost:
		mock = mocha.Post(pathMatcher)
	case http.MethodDelete:
		mock = mocha.Delete(pathMatcher)
	default:
		t.Fatalf("unsupported HTTP verb: %s", request.verb)
	}

	if request.headers != nil {
		for key, value := range request.headers {
			mock = mock.Header(key, expect.ToEqual(value))
		}
	}

	replySequence := reply.Seq()
	for _, response := range responses {
		responseStatus := response.statusCode
		if responseStatus == 0 {
			responseStatus = http.StatusOK
		}

		if request.bodyString != "" {
			mock = mock.Body(expect.Func(func(v any, _ expect.Args) (bool, error) {
				bodyRaw, err := json.Marshal(v)
				if err != nil {
					return false, errors.New("unexpected error to read request body on mocha")
				}
				require.Equal(t, request.bodyString, string(bodyRaw))
				return true, nil
			}))
		}

		replySequence.Add(reply.Status(responseStatus).Header("content-type", "application/json").BodyJSON(response.body))
	}
	mock = mock.Reply(replySequence)

	m.AddMocks(mock)

	return m
}

type mockedTokenManager struct{}

func (t *mockedTokenManager) SetAuthHeader(req *http.Request) error {
	return nil
}
