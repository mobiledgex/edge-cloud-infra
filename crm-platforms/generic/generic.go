package generic

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type GenericPlatform struct {
	openRCVars   map[string]string
	VMProperties *vmlayer.VMProperties
	TestMode     bool
	caches       *platform.Caches
}

func (o *GenericPlatform) GetType() string {
	return "generic"
}

func (o *GenericPlatform) SetVMProperties(vmProperties *vmlayer.VMProperties) {
	o.VMProperties = vmProperties
}

func (o *GenericPlatform) GetCloudletKey() *edgeproto.CloudletKey {
	return o.VMProperties.CommonPf.PlatformConfig.CloudletKey
}

func (o *GenericPlatform) InitProvider(ctx context.Context, caches *platform.Caches, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitProvider for VSphere")
	o.SetCaches(ctx, caches)
	return nil
}

func (o *GenericPlatform) SetCaches(ctx context.Context, caches *platform.Caches) {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetCaches")
	o.caches = caches
}

func (o *GenericPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo ")
	var err error
	info.Flavors, err = o.GetFlavorList(ctx)
	return err
}

// alphanumeric plus -_. first char must be alpha, <= 255 chars.
func (o *GenericPlatform) NameSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "",
		",", "",
		"!", "")
	str := r.Replace(name)
	if str == "" {
		return str
	}
	if !unicode.IsLetter(rune(str[0])) {
		// first character must be alpha
		str = "a" + str
	}
	if len(str) > 255 {
		str = str[:254]
	}
	return str
}

// Generic IdSanitize is the same as NameSanitize
func (o *GenericPlatform) IdSanitize(name string) string {
	return o.NameSanitize(name)
}

func (o *GenericPlatform) GetResourceID(ctx context.Context, resourceType vmlayer.ResourceType, resourceName string) (string, error) {
	switch resourceType {
	case vmlayer.ResourceTypeSecurityGroup:
		// not exists, just return same value
		return resourceName, nil
	}
	return "", fmt.Errorf("GetResourceID not implemented for resource type: %s ", resourceType)
}
