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
			Name:         "CreateSelfFederation",
			Use:          "create",
			Short:        "Create Self Federation",
			RequiredArgs: strings.Join(FederationRequiredArgs, " "),
			OptionalArgs: strings.Join(FederationOptionalArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.OperatorFederation{},
			Path:         "/auth/federation/self/create",
		},
		&ApiCommand{
			Name:         "CreatePartnerFederation",
			Use:          "create",
			Short:        "Create Partner Federation",
			RequiredArgs: strings.Join(FederationRequiredArgs, " "),
			Comments:     FederationComments,
			ReqData:      &ormapi.OperatorFederation{},
			ReplyData:    &ormapi.Result{},
			Path:         "/auth/federation/partner/create",
		},
	}
	AllApis.AddGroup(FederationGroup, "Federation APIs", cmds)
}

var FederationRequiredArgs = []string{
	"federationid",
	"operatorid",
	"countrycode",
	"mcc",
	"mncs",
}

var FederationOptionalArgs = []string{
	"locatorendpoint",
}

var FederationPartnerRequiredArgs = []string{
	"federationid",
	"federationAddr",
	"operatorid",
	"countrycode",
}

var FederationComments = map[string]string{
	"federationid":    "Globally unique string used to authenticate operations over federation interface",
	"federationAddr":  "Federation access point address",
	"operatorid":      "Globally unique string to identify an operator gMEC",
	"countrycode":     "ISO 3166-1 Alpha-2 code for the country where operator gMEC is located",
	"mcc":             "Mobile country code of operator sending the request",
	"mncs":            "Comma separated list of mobile network codes of operator sending the request",
	"locatorendpoint": "IP and Port of discovery service URL of gMEC",
}
