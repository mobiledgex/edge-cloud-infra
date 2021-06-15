package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

const RateLimitSettingsMcGroup = "RateLimitSettingsMc"

var RateLimitSettingsMcRequiredArgs = []string{
	"apiname",
	"ratelimittarget",
}

func init() {
	cmds := []*ApiCommand{&ApiCommand{
		Name:         "UpdateRateLimitSettingsMc",
		Use:          "update",
		Short:        "Update master controller ratelimitsettings",
		RequiredArgs: strings.Join(RateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(RateLimitSettingsOptionalArgs, " "),
		Comments:     RateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitSettings{},
		Path:         "/auth/ratelimitsettingsmc/update",
	}, &ApiCommand{
		Name:         "DeleteRateLimitSettingsMc",
		Use:          "delete",
		Short:        "Delete master controller ratelimitsettings",
		RequiredArgs: strings.Join(RateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(RateLimitSettingsOptionalArgs, " "),
		Comments:     RateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitSettings{},
		Path:         "/auth/ratelimitsettingsmc/delete",
	}, &ApiCommand{
		Name:         "CreateRateLimitSettingsMc",
		Use:          "create",
		Short:        "Create master controller ratelimitsettings",
		RequiredArgs: strings.Join(RateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(RateLimitSettingsOptionalArgs, " "),
		Comments:     RateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitSettings{},
		Path:         "/auth/ratelimitsettingsmc/create",
	}, &ApiCommand{
		Name:         "ShowRateLimitSettingsMc",
		Use:          "show",
		Short:        "Show master controller ratelimitsettings",
		OptionalArgs: strings.Join(append(RateLimitSettingsMcRequiredArgs, RateLimitSettingsOptionalArgs...), " "),
		Comments:     RateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitSettings{},
		ReplyData:    &[]ormapi.McRateLimitSettings{},
		Path:         "/auth/ratelimitsettingsmc/show",
	}}
	AllApis.AddGroup(RateLimitSettingsMcGroup, "Manage global ratelimitsettings", cmds)
}
