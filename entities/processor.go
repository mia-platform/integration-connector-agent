package entities

type Processor interface {
	Process(data PipelineEvent) (PipelineEvent, error)
}
