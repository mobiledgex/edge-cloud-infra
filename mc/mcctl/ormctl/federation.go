package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

const (
	FederationGroup = "Federation"
)

func init() {
	cmds := []*ApiCommand{
		&ApiCommand{
			Name:         "CreateSelfFederator",
			Use:          "create",
			Short:        "Create Self Federator",
			RequiredArgs: "operatorid countrycode mcc mncs regions",
			OptionalArgs: "locatorendpoint",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Federator{},
			Path:         "/auth/federator/self/create",
		},
		&ApiCommand{
			Name:         "UpdateSelfFederator",
			Use:          "update",
			Short:        "Update Self Federator",
			OptionalArgs: "mcc mncs regions locatorendpoint",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/update",
		},
		&ApiCommand{
			Name:      "DeleteSelfFederator",
			Use:       "delete",
			Short:     "Delete Self Federator",
			Comments:  FederatorComments,
			ReqData:   &ormapi.FederatorRequest{},
			ReplyData: &ormapi.Result{},
			Path:      "/auth/federator/self/delete",
		},
		&ApiCommand{
			Name:         "ShowSelfFederator",
			Use:          "show",
			Short:        "Show Self Federator",
			OptionalArgs: strings.Join(FederatorArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &[]ormapi.FederatorRequest{},
			Path:         "/auth/federator/self/show",
		},
		&ApiCommand{
			Name:         "AddPartnerFederator",
			Use:          "addpartner",
			Short:        "Add Partner Federator",
			RequiredArgs: "federationid federationaddr operatorid countrycode",
			PasswordArg:  "federationid",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorPartnerRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/add",
		},
		&ApiCommand{
			Name:         "RemovePartnerFederator",
			Use:          "removepartner",
			Short:        "Remove Partner Federator",
			RequiredArgs: "federationid",
			PasswordArg:  "federationid",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorPartnerRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/remove",
		},
		&ApiCommand{
			Name:         "ShowPartnerFederator",
			Use:          "showpartner",
			Short:        "Show Partner Federator",
			OptionalArgs: strings.Join(FederatorArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &[]ormapi.FederatorRequest{},
			Path:         "/auth/federator/partner/show",
		},
		&ApiCommand{
			Name:         "ShowPartnerFederatorRole",
			Use:          "showpartnerrole",
			Short:        "Show Partner Federator Role",
			OptionalArgs: strings.Join(FederatorArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federator{},
			ReplyData:    &[]ormapi.FederatorRoleResponse{},
			Path:         "/auth/federator/partner/show/role",
		},
		&ApiCommand{
			Name:         "CreateSelfFederatorZone",
			Use:          "createzone",
			Short:        "Create Self Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			OptionalArgs: strings.Join(FederatorZoneOptionalArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/create",
		},
		&ApiCommand{
			Name:         "DeleteSelfFederatorZone",
			Use:          "deletezone",
			Short:        "Delete Self Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/delete",
		},
		&ApiCommand{
			Name:         "ShareSelfFederatorZone",
			Use:          "sharezone",
			Short:        "Share Self Federation Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/share",
		},
		&ApiCommand{
			Name:         "UnshareSelfFederationZone",
			Use:          "unsharezone",
			Short:        "Unshare Self Federation Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/unshare",
		},
		&ApiCommand{
			Name:         "RegisterPartnerFederationZone",
			Use:          "registerzone",
			Short:        "Register Partner Federation Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/register",
		},
		&ApiCommand{
			Name:         "DeRegisterPartnerFederationZone",
			Use:          "deregisterzone",
			Short:        "DeRegister Partner Federation Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/deregister",
		},
		&ApiCommand{
			Name:         "ShowFederationZone",
			Use:          "showzone",
			Short:        "Show Federation Zones",
			OptionalArgs: strings.Join(append(FederatorZoneRequiredArgs, FederatorZoneOptionalArgs...), " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &[]ormapi.FederatorZoneDetails{},
			Path:         "/auth/federator/zone/show",
		},
	}
	AllApis.AddGroup(FederationGroup, "Federation APIs", cmds)
}

var FederatorArgs = []string{
	"federationid",
	"operatorid",
	"countrycode",
	"mcc",
	"mncs",
	"regions",
	"locatorendpoint",
}

var FederatorComments = map[string]string{
	"federationid":    "Globally unique string used to authenticate operations over federation interface",
	"federationaddr":  "Federation access point address",
	"operatorid":      "Globally unique string to identify an operator platform",
	"countrycode":     "ISO 3166-1 Alpha-2 code for the country where operator platform is located",
	"mcc":             "Mobile country code of operator sending the request",
	"mncs":            "List of mobile network codes of operator sending the request",
	"regions":         "List of regions all the zone cloudlets belongs to",
	"locatorendpoint": "IP and Port of discovery service URL of operator platform",
}

var FederatorZoneRequiredArgs = []string{
	"zoneid",
}

var FederatorZoneOptionalArgs = []string{
	"federationid",
	"geolocation",
	"city",
	"state",
	"locality",
	"cloudlets",
}

var FederatorZoneComments = map[string]string{
	"zoneid":       "Globally unique string used to authenticate operations over federation interface",
	"federationid": "Globally unique string used to authenticate operations over federation interface",
	"geolocation":  "GPS co-ordinates associated with the zone (in decimal format)",
	"city":         "Comma seperated list of cities under this zone",
	"state":        "Comma seperated list of states under this zone",
	"locality":     "Type of locality eg rural, urban etc.",
	"cloudlets":    "List of cloudlets belonging to the federation zone",
}
