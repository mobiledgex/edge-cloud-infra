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
			RequiredArgs: strings.Join(FederatorRequiredArgs, " "),
			OptionalArgs: strings.Join(FederatorOptionalArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federator{},
			ReplyData:    &ormapi.Federator{},
			Path:         "/auth/federator/self/create",
		},
		&ApiCommand{
			Name:         "UpdateSelfFederator",
			Use:          "update",
			Short:        "Update Self Federator",
			RequiredArgs: "operatorid countrycode",
			OptionalArgs: "mcc mncs locatorendpoint",
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federator{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/update",
		},
		&ApiCommand{
			Name:         "DeleteSelfFederator",
			Use:          "delete",
			Short:        "Delete Self Federator",
			RequiredArgs: "operatorid countrycode",
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federator{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/delete",
		},
		&ApiCommand{
			Name:         "ShowSelfFederator",
			Use:          "showselffederator",
			Short:        "Show Self Federator",
			RequiredArgs: "operatorid countrycode",
			OptionalArgs: strings.Join(append(FederatorRequiredArgs, FederatorOptionalArgs...), " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federator{},
			ReplyData:    &[]ormapi.Federator{},
			Path:         "/auth/federator/self/show",
		},
		&ApiCommand{
			Name:         "CreateSelfFederatorZone",
			Use:          "createzone",
			Short:        "Create Self Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequiredArgs, " "),
			OptionalArgs: strings.Join(FederatorZoneOptionalArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/create",
		},
		&ApiCommand{
			Name:         "DeleteSelfFederatorZone",
			Use:          "deletezone",
			Short:        "Delete Self Federator Zone",
			RequiredArgs: "zoneid operatorid countrycode",
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/delete",
		},
		&ApiCommand{
			Name:         "ShareSelfFederatorZone",
			Use:          "sharezone",
			Short:        "Share Self Federator Zone",
			RequiredArgs: strings.Join(FederatedSelfZoneArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedSelfZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/share",
		},
		&ApiCommand{
			Name:         "UnshareSelfFederatorZone",
			Use:          "unsharezone",
			Short:        "Unshare Self Federator Zone",
			RequiredArgs: strings.Join(FederatedSelfZoneArgs, " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedSelfZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/unshare",
		},
		&ApiCommand{
			Name:         "CreatePartnerFederator",
			Use:          "createpartner",
			Short:        "Create Partner Federator",
			RequiredArgs: strings.Join(append(SelfFederatorArgs, FederationRequiredArgs...), " "),
			OptionalArgs: strings.Join(FederatorOptionalArgs, " "),
			PasswordArg:  "federationkey",
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/create",
		},
		&ApiCommand{
			Name:         "DeletePartnerFederator",
			Use:          "deletepartner",
			Short:        "Delete Partner Federator",
			RequiredArgs: "selfoperatorid selfcountrycode operatorid countrycode",
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/delete",
		},
		&ApiCommand{
			Name:         "RegisterPartnerFederatorZone",
			Use:          "registerzone",
			Short:        "Register Partner Federator Zone",
			RequiredArgs: strings.Join(append(SelfFederatorArgs, FederatorZoneRequiredArgs...), " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedPartnerZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/register",
		},
		&ApiCommand{
			Name:         "DeRegisterPartnerFederatorZone",
			Use:          "deregisterzone",
			Short:        "DeRegister Partner Federator Zone",
			RequiredArgs: "selfoperatorid selfcountrycode operatorid countrycode zoneid",
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedPartnerZone{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/deregister",
		},
		&ApiCommand{
			Name:         "CreateFederation",
			Use:          "createfederation",
			Short:        "Create Federation",
			RequiredArgs: strings.Join(FederationArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/create",
		},
		&ApiCommand{
			Name:         "DeleteFederation",
			Use:          "deletefederation",
			Short:        "Delete Federation",
			RequiredArgs: strings.Join(FederationArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/delete",
		},
		&ApiCommand{
			Name:         "ShowFederation",
			Use:          "showfederation",
			Short:        "Show Federation",
			RequiredArgs: strings.Join(SelfFederatorArgs, " "),
			OptionalArgs: strings.Join(FederationRequiredArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.Federation{},
			ReplyData:    &[]ormapi.Federation{},
			Path:         "/auth/federation/show",
		},
		&ApiCommand{
			Name:         "ShowFederatedSelfZone",
			Use:          "showfederatedselfzone",
			Short:        "Show Federated Self Zones",
			RequiredArgs: strings.Join(SelfFederatorArgs, " "),
			OptionalArgs: "partneroperatorid partnercountrycode zoneid",
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedSelfZone{},
			ReplyData:    &[]ormapi.FederatedSelfZone{},
			Path:         "/auth/federation/self/zone/show",
		},
		&ApiCommand{
			Name:         "ShowFederatedPartnerZone",
			Use:          "showfederatedpartnerzone",
			Short:        "Show Federated Partner Zones",
			RequiredArgs: strings.Join(SelfFederatorArgs, " "),
			OptionalArgs: "operatorid countrycode zoneid",
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatedPartnerZone{},
			ReplyData:    &[]ormapi.FederatedPartnerZone{},
			Path:         "/auth/federation/partner/zone/show",
		},
	}
	AllApis.AddGroup(FederationGroup, "Federation APIs", cmds)
}

var SelfFederatorArgs = []string{
	"selfoperatorid",
	"selfcountrycode",
}

var FederatorRequiredArgs = []string{
	"operatorid",
	"countrycode",
	"mcc",
	"mnc",
}

var FederatorOptionalArgs = []string{
	"federationkey",
	"locatorendpoint",
}

var FederationRequiredArgs = []string{
	"operatorid",
	"countrycode",
	"mcc",
	"mnc",
	"federationkey",
	"federationaddr",
}

var FederationArgs = []string{
	"selfoperatorid",
	"selfcountrycode",
	"partneroperatorid",
	"partnercountrycode",
}

var FederatorZoneRequiredArgs = []string{
	"zoneid",
	"operatorid",
	"countrycode",
	"cloudlets",
	"geolocation",
}

var FederatorZoneOptionalArgs = []string{
	"city",
	"state",
	"locality",
}

var FederatedSelfZoneArgs = []string{
	"zoneid",
	"selfoperatorid",
	"selfcountrycode",
	"partneroperatorid",
	"partnercountrycode",
}

var FederatorComments = map[string]string{
	"selfoperatorid":     "Self federator operator ID",
	"selfcountrycode":    "Self federator country code",
	"partneroperatorid":  "Partner federator operator ID",
	"partnercountrycode": "Partner federator country code",
	"operatorid":         "Globally unique string to identify an operator platform",
	"countrycode":        "ISO 3166-1 Alpha-2 code for the country where operator platform is located",
	"federationkey":      "Globally unique string used to authenticate operations over federation interface",
	"federationaddr":     "Federation access point address",
	"mcc":                "Mobile country code of operator sending the request",
	"mnc":                "List of mobile network codes of operator sending the request",
	"locatorendpoint":    "IP and Port of discovery service URL of operator platform",
}

var FederatorZoneComments = map[string]string{
	"operatorid":         "Globally unique string to identify an operator platform",
	"countrycode":        "ISO 3166-1 Alpha-2 code for the country where operator platform is located",
	"zoneid":             "Globally unique string used to authenticate operations over federation interface",
	"geolocation":        "GPS co-ordinates associated with the zone (in decimal format)",
	"city":               "Comma seperated list of cities under this zone",
	"state":              "Comma seperated list of states under this zone",
	"locality":           "Type of locality eg rural, urban etc.",
	"region":             "Regions in which cloudlet belongs to",
	"cloudlets":          "List of cloudlets belonging to the federator zone",
	"selfoperatorid":     "Self federator operator ID",
	"selfcountrycode":    "Self federator country code",
	"partneroperatorid":  "Partner federator operator ID",
	"partnercountrycode": "Partner federator country code",
}
