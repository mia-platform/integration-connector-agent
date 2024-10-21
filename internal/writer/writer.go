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

package writer

import "context"

// Writer interface abstract the implementation of an integration pipeline target. The concrete implementation has
// to know how to write and delete a Data.
type Writer[Data any] interface {
	// Write will save the Data to the destination configured in the Writer. Writer implementation can choose to
	// implement this function as a single write or to update data based on an identifier
	Write(context.Context, Data) error

	// Delete will delete the Data from the destination configured in the Writer.
	Delete(context.Context, Data) error
}
