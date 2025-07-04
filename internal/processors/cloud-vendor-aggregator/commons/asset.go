package commons

const (
	AWSAssetProvider   = "aws"
	AzureAssetProvider = "azure"
	GCPAssetProvider   = "gcp"
)

type Tags map[string]string

type Asset struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Provider      string   `json:"provider"`
	Location      string   `json:"location"`
	Relationships []string `json:"relationships"`
	Tags          Tags     `json:"tags"`
	RawData       []byte   `json:"rawData"`
}
