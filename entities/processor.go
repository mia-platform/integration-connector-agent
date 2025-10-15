// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package entities

type Processor interface {
	Process(data PipelineEvent) (PipelineEvent, error)
}

type Initializable interface {
	Init(config []byte) error
}

type InitializableProcessor interface {
	Processor
	Initializable
}
