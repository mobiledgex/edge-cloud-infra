package infracommon

import (
	"fmt"

	yaml "github.com/mobiledgex/yaml/v2"
)

type ManifestBodyType string

const ManifestURL ManifestBodyType = "URL"
const ManifestCode ManifestBodyType = "code"
const ManifestText ManifestBodyType = "text"
const ManifestCommand ManifestBodyType = "command"

type CloudletManifestItem struct {
	Title       string
	BodyType    ManifestBodyType
	BodyContent string
}
type CloudletManifest struct {
	ManifestItems []CloudletManifestItem
}

func (m *CloudletManifest) ToString() (string, error) {
	out, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal manifest: %v", err)
	}
	return string(out), nil
}
