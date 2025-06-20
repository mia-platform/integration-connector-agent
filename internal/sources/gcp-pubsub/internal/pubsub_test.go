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

package internal

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	log, _ := test.NewNullLogger()
	c, err := New(t.Context(), log, PubSubConfig{
		ProjectID:          "test-project",
		AckDeadlineSeconds: 10,
		TopicName:          "test-topic",
		SubscriptionID:     "test-subscription",
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}
