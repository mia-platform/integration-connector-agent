// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package entities

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Write-0]
	_ = x[Delete-1]
}

//nolint:revive // Generated code
const _Operation_name = "WriteDelete"

//nolint:revive // Generated code
var _Operation_index = [...]uint8{0, 5, 11}

func (i Operation) String() string {
	if i < 0 || i >= Operation(len(_Operation_index)-1) {
		return "Operation(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Operation_name[_Operation_index[i]:_Operation_index[i+1]]
}
