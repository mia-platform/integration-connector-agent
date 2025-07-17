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

package consolecatalog

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

var dynamicOperatorRegexp = regexp.MustCompile(`{{\s*(.*?)\s*}}`)

func templetize(template string, raw []byte) (string, error) {
	matches := dynamicOperatorRegexp.FindAllStringSubmatch(template, -1)

	var result = template
	for _, match := range matches {
		if len(match) < 2 {
			return "", fmt.Errorf("invalid template: %s", template)
		}

		key := match[1]
		res := gjson.Get(string(raw), key)
		result = strings.ReplaceAll(result, match[0], res.String())
	}

	return result, nil
}

func slugify(input string) (output string) {
	reg := regexp.MustCompile(`[^a-z0-9_]+`)

	output = strings.ToLower(input)
	output = reg.ReplaceAllString(output, "-")
	output = strings.Trim(output, "-")
	return
}
