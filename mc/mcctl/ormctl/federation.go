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
			Name:         "CreateFederation",
			Use:          "create",
			Short:        "Create Federation",
			RequiredArgs: "operatorid countrycode mcc mncs",
			OptionalArgs: "locatorendpoint",
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.OperatorFederation{},
			Path:         "/auth/federation/create",
		},
		&ApiCommand{
			Name:         "UpdateFederation",
			Use:          "update",
			Short:        "Update Federation",
			OptionalArgs: "mcc mncs locatorendpoint",
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/update",
		},
		&ApiCommand{
			Name:      "DeleteFederation",
			Use:       "delete",
			Short:     "Delete Federation",
			Comments:  FederationComments,
			ReqData:   &ormapi.OperatorFederation{},
			ReplyData: &ormapi.Result{},
			Path:      "/auth/federation/delete",
		},
		&ApiCommand{
			Name:         "ShowFederation",
			Use:          "show",
			Short:        "Show Federation",
			OptionalArgs: strings.Join(FederationArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &[]ormapi.OperatorFederation{},
			Path:         "/auth/federation/show",
		},
		&ApiCommand{
			Name:         "AddFederationPartner",
			Use:          "addpartner",
			Short:        "Add Federation Partner",
			RequiredArgs: "federationid federationaddr operatorid countrycode",
			PasswordArg:  "federationid",
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/partner/add",
		},
		&ApiCommand{
			Name:         "RemoveFederationPartner",
			Use:          "removepartner",
			Short:        "Remove Federation Partner",
			RequiredArgs: "federationid",
			PasswordArg:  "federationid",
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/partner/remove",
		},
		&ApiCommand{
			Name:         "CreateFederationZone",
			Use:          "createzone",
			Short:        "Create Federation Zone",
			RequiredArgs: strings.Join(FederationZoneRequiredArgs, " "),
			OptionalArgs: strings.Join(FederationZoneOptionalArgs, " "),
			Comments:     FederationZoneComments,
			ReqData:      &ormapi.OperatorZoneCloudletMap{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/zone/create",
		},
		&ApiCommand{
			Name:         "DeleteFederationZone",
			Use:          "deletezone",
			Short:        "Delete Federation Zone",
			RequiredArgs: strings.Join(FederationZoneRequiredArgs, " "),
			Comments:     FederationZoneComments,
			ReqData:      &ormapi.OperatorZoneCloudletMap{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/zone/delete",
		},
		&ApiCommand{
			Name:         "RegisterFederationZone",
			Use:          "registerzone",
			Short:        "Register Federation Zone",
			RequiredArgs: strings.Join(FederationZoneRequiredArgs, " "),
			Comments:     FederationZoneComments,
			ReqData:      &ormapi.OperatorZoneCloudletMap{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/zone/register",
		},
		&ApiCommand{
			Name:         "DeRegisterFederationZone",
			Use:          "deregisterzone",
			Short:        "DeRegister Federation Zone",
			RequiredArgs: strings.Join(FederationZoneRequiredArgs, " "),
			Comments:     FederationZoneComments,
			ReqData:      &ormapi.OperatorZoneCloudletMap{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/zone/deregister",
		},
		&ApiCommand{
			Name:         "ShowFederationZone",
			Use:          "showzone",
			Short:        "Show Federation Zones",
			OptionalArgs: strings.Join(append(FederationZoneRequiredArgs, FederationZoneOptionalArgs...), " "),
			Comments:     FederationZoneComments,
			ReqData:      &ormapi.OperatorZoneCloudletMap{},
			ReplyData:    &[]ormapi.OperatorZoneCloudletMap{},
			Path:         "/auth/federation/zone/show",
		},
	}
	AllApis.AddGroup(FederationGroup, "Federation APIs", cmds)
}

var FederationArgs = []string{
	"federationid",
	"operatorid",
	"countrycode",
	"mcc",
	"mncs",
	"locatorendpoint",
}

var FederationComments = map[string]string{
	"federationid":    "Globally unique string used to authenticate operations over federation interface",
	"federationaddr":  "Federation access point address",
	"operatorid":      "Globally unique string to identify an operator gMEC",
	"countrycode":     "ISO 3166-1 Alpha-2 code for the country where operator gMEC is located",
	"mcc":             "Mobile country code of operator sending the request",
	"mncs":            "Comma separated list of mobile network codes of operator sending the request",
	"locatorendpoint": "IP and Port of discovery service URL of gMEC",
}

var FederationZoneRequiredArgs = []string{
	"zoneid",
}

var FederationZoneOptionalArgs = []string{
	"federationid",
	"geolocation",
	"city",
	"state",
	"locality",
	"cloudlets",
}

var FederationZoneComments = map[string]string{
	"zoneid":       "Globally unique string used to authenticate operations over federation interface",
	"federationid": "Globally unique string used to authenticate operations over federation interface",
	"geolocation":  "GPS co-ordinates associated with the zone (in decimal format)",
	"city":         "Comma seperated list of cities under this zone",
	"state":        "Comma seperated list of states under this zone",
	"locality":     "Type of locality eg rural, urban etc.",
	"cloudlets":    "List of cloudlets belonging to the federation zone",
}
