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
			RequiredArgs: strings.Join(FederatorArgs, " "),
			OptionalArgs: strings.Join(FederatorOptionalArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Federator{},
			Path:         "/auth/federator/self/create",
		},
		&ApiCommand{
			Name:         "UpdateSelfFederator",
			Use:          "update",
			Short:        "Update Self Federator",
			OptionalArgs: "mcc mncs locatorendpoint",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/update",
		},
		&ApiCommand{
			Name:         "DeleteSelfFederator",
			Use:          "delete",
			Short:        "Delete Self Federator",
			RequiredArgs: "operatorid countrycode",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/delete",
		},
		&ApiCommand{
			Name:         "ShowSelfFederator",
			Use:          "showselffederator",
			Short:        "Show Self Federator",
			OptionalArgs: strings.Join(FederatorArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &[]ormapi.FederatorRequest{},
			Path:         "/auth/federator/self/show",
		},
		&ApiCommand{
			Name:         "CreatePartnerFederator",
			Use:          "createpartner",
			Short:        "Create Partner Federator",
			RequiredArgs: strings.Join(append(SelfFederatorArgs, FederatorArgs...), " "),
			OptionalArgs: strings.Join(FederatorOptionalArgs, " "),
			PasswordArg:  "federationkey",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/create",
		},
		&ApiCommand{
			Name:         "DeletePartnerFederator",
			Use:          "deletepartner",
			Short:        "Delete Partner Federator",
			RequiredArgs: "selfoperatorid selfcountrycode operatorid countrycode",
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/delete",
		},
		&ApiCommand{
			Name:         "ShowPartnerFederator",
			Use:          "showpartnerfederator",
			Short:        "Show Partner Federator",
			OptionalArgs: strings.Join(FederatorArgs, " "),
			Comments:     FederatorComments,
			ReqData:      &ormapi.FederatorRequest{},
			ReplyData:    &[]ormapi.FederatorRequest{},
			Path:         "/auth/federator/partner/show",
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
			Short:        "Share Self Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequestArgs, " "),
			Comments:     FederatorZoneShareComments,
			ReqData:      &ormapi.FederatorZoneRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/share",
		},
		&ApiCommand{
			Name:         "UnshareSelfFederatorZone",
			Use:          "unsharezone",
			Short:        "Unshare Self Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequestArgs, " "),
			Comments:     FederatorZoneShareComments,
			ReqData:      &ormapi.FederatorZoneRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/self/zone/unshare",
		},
		&ApiCommand{
			Name:         "RegisterPartnerFederatorZone",
			Use:          "registerzone",
			Short:        "Register Partner Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequestArgs, " "),
			Comments:     FederatorZoneRegisterComments,
			ReqData:      &ormapi.FederatorZoneRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/register",
		},
		&ApiCommand{
			Name:         "DeRegisterPartnerFederatorZone",
			Use:          "deregisterzone",
			Short:        "DeRegister Partner Federator Zone",
			RequiredArgs: strings.Join(FederatorZoneRequestArgs, " "),
			Comments:     FederatorZoneRegisterComments,
			ReqData:      &ormapi.FederatorZoneRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federator/partner/zone/deregister",
		},
		&ApiCommand{
			Name:         "ShowFederatorZone",
			Use:          "showzone",
			Short:        "Show Federator Zones",
			OptionalArgs: strings.Join(append(FederatorZoneRequiredArgs, FederatorZoneOptionalArgs...), " "),
			Comments:     FederatorZoneComments,
			ReqData:      &ormapi.FederatorZoneDetails{},
			ReplyData:    &[]ormapi.FederatorZoneDetails{},
			Path:         "/auth/federator/zone/show",
		},
		&ApiCommand{
			Name:         "CreateFederation",
			Use:          "createfederation",
			Short:        "Create Federation",
			RequiredArgs: strings.Join(FederationArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.FederationRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/create",
		},
		&ApiCommand{
			Name:         "DeleteFederation",
			Use:          "deletefederation",
			Short:        "Delete Federation",
			RequiredArgs: strings.Join(FederationArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.FederationRequest{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/delete",
		},
		&ApiCommand{
			Name:         "ShowFederation",
			Use:          "showfederation",
			Short:        "Show Federation",
			OptionalArgs: strings.Join(FederationArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.FederationRequest{},
			ReplyData:    &[]ormapi.Federation{},
			Path:         "/auth/federation/show",
		},
	}
	AllApis.AddGroup(FederationGroup, "Federation APIs", cmds)
}

var SelfFederatorArgs = []string{
	"selfoperatorid",
	"selfcountrycode",
}

var FederatorArgs = []string{
	"operatorid",
	"countrycode",
	"mcc",
	"mncs",
}

var FederatorOptionalArgs = []string{
	"federationkey",
	"locatorendpoint",
}

var FederatorComments = map[string]string{
	"selfoperatorid":  "Self federator operator ID",
	"selfcountrycode": "Self federator country code",
	"federationkey":   "Globally unique string used to authenticate operations over federation interface",
	"federationaddr":  "Federation access point address",
	"operatorid":      "Globally unique string to identify an operator platform",
	"countrycode":     "ISO 3166-1 Alpha-2 code for the country where operator platform is located",
	"mcc":             "Mobile country code of operator sending the request",
	"mncs":            "List of mobile network codes of operator sending the request",
	"locatorendpoint": "IP and Port of discovery service URL of operator platform",
}

var FederationArgs = []string{
	"selfoperatorid",
	"selfcountrycode",
	"partneroperatorid",
	"partnercountrycode",
}

var FederationComments = map[string]string{
	"selfoperatorid":     "Self federator operator ID",
	"selfcountrycode":    "Self federator country code",
	"partneroperatorid":  "Partner federator operator ID",
	"partnercountrycode": "Partner federator country code",
}

var FederatorZoneRequiredArgs = []string{
	"zoneid",
}

var FederatorZoneOptionalArgs = []string{
	"geolocation",
	"city",
	"state",
	"locality",
	"cloudlets",
}

var FederatorZoneComments = map[string]string{
	"zoneid":      "Globally unique string used to authenticate operations over federation interface",
	"geolocation": "GPS co-ordinates associated with the zone (in decimal format)",
	"city":        "Comma seperated list of cities under this zone",
	"state":       "Comma seperated list of states under this zone",
	"locality":    "Type of locality eg rural, urban etc.",
	"cloudlets":   "List of cloudlets belonging to the federator zone",
}

var FederatorZoneRequestArgs = []string{
	"zoneid",
	"selfoperatorid",
	"selfcountrycode",
	"partneroperatorid",
	"partnercountrycode",
}

var FederatorZoneShareComments = map[string]string{
	"zoneid":             "Unique ID to identify zone",
	"selfoperatorid":     "Self federator operator ID",
	"selfcountrycode":    "Self federator country code",
	"partneroperatorid":  "Partner federator operator ID",
	"partnercountrycode": "Partner federator country code",
}

var FederatorZoneRegisterComments = map[string]string{
	"zoneid":             "Unique ID to identify zone",
	"selfoperatorid":     "Self federator operator ID",
	"selfcountrycode":    "Self federator country code",
	"partneroperatorid":  "Partner federator operator ID",
	"partnercountrycode": "Partner federator country code",
}
