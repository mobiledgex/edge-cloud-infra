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

var FlowRateLimitSettingsMcRequiredArgs = []string{
	"flowsettingsname",
	"apiname",
	"ratelimittarget",
}

var MaxReqsRateLimitSettingsMcRequiredArgs = []string{
	"maxreqssettingsname",
	"apiname",
	"ratelimittarget",
}

func init() {
	cmds := []*ApiCommand{&ApiCommand{
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
	}, &ApiCommand{
		Name:         "CreateFlowRateLimitSettingsMc",
		Use:          "createflow",
		Short:        "Create master controller flowratelimitsettings",
		RequiredArgs: strings.Join(FlowRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(FlowRateLimitSettingsOptionalArgs, " "),
		Comments:     FlowRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitFlowSettings{},
		Path:         "/auth/ratelimitsettingsmc/createflow",
	}, &ApiCommand{
		Name:         "UpdateFlowRateLimitSettingsMc",
		Use:          "updateflow",
		Short:        "Update master controller flowratelimitsettings",
		RequiredArgs: strings.Join(FlowRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(FlowRateLimitSettingsOptionalArgs, " "),
		Comments:     FlowRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitFlowSettings{},
		Path:         "/auth/ratelimitsettingsmc/updateflow",
	}, &ApiCommand{
		Name:         "DeleteFlowRateLimitSettingsMc",
		Use:          "deleteflow",
		Short:        "Delete master controller flowratelimitsettings",
		RequiredArgs: strings.Join(FlowRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(FlowRateLimitSettingsOptionalArgs, " "),
		Comments:     FlowRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitFlowSettings{},
		Path:         "/auth/ratelimitsettingsmc/deleteflow",
	}, &ApiCommand{
		Name:         "CreateMaxReqsRateLimitSettingsMc",
		Use:          "createmaxreqs",
		Short:        "Create master controller maxreqsratelimitsettings",
		RequiredArgs: strings.Join(MaxReqsRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(MaxReqsRateLimitSettingsOptionalArgs, " "),
		Comments:     MaxReqsRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitMaxReqsSettings{},
		Path:         "/auth/ratelimitsettingsmc/createmaxreqs",
	}, &ApiCommand{
		Name:         "UpdateMaxReqsRateLimitSettingsMc",
		Use:          "updatemaxreqs",
		Short:        "Update master controller maxreqsratelimitsettings",
		RequiredArgs: strings.Join(MaxReqsRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(MaxReqsRateLimitSettingsOptionalArgs, " "),
		Comments:     MaxReqsRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitMaxReqsSettings{},
		Path:         "/auth/ratelimitsettingsmc/updatemaxreqs",
	}, &ApiCommand{
		Name:         "DeleteMaxReqsRateLimitSettingsMc",
		Use:          "deletemaxreqs",
		Short:        "Delete master controller maxreqsratelimitsettings",
		RequiredArgs: strings.Join(MaxReqsRateLimitSettingsMcRequiredArgs, " "),
		OptionalArgs: strings.Join(MaxReqsRateLimitSettingsOptionalArgs, " "),
		Comments:     MaxReqsRateLimitSettingsComments,
		ReqData:      &ormapi.McRateLimitMaxReqsSettings{},
		Path:         "/auth/ratelimitsettingsmc/deletemaxreqs",
	}}
	AllApis.AddGroup(RateLimitSettingsMcGroup, "Manage global ratelimitsettings", cmds)
}
