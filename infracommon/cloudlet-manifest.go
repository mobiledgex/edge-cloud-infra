package infracommon

import (
	"fmt"

	yaml "github.com/mobiledgex/yaml/v2"
)

type ManifestBodyType string

const ManifestNone ManifestBodyType = "none"
const ManifestURL ManifestBodyType = "url"
const ManifestCode ManifestBodyType = "code"

type CloudletManifestItem struct {
	Title       string
	ContentType ManifestBodyType
	Content     string
}
type CloudletManifest struct {
	ManifestItems []CloudletManifestItem
}

func (m *CloudletManifest) AddItem(title string, bodyType ManifestBodyType, content string) {
	item := CloudletManifestItem{
		Title:       title,
		ContentType: bodyType,
		Content:     content,
	}
	m.ManifestItems = append(m.ManifestItems, item)
}

func (m *CloudletManifest) ToString() (string, error) {
	out, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal manifest: %v", err)
	}
	return string(out), nil
}
