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

package rpcprocessor

import (
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	"github.com/sirupsen/logrus"
)

type Logger = *logrus.Logger

func NewLogger(level string) (*logrus.Logger, error) {
	l, err := glogrus.InitHelper(glogrus.InitOptions{
		Level: level,
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}
