package ormctl

import (
	"github.com/mobiledgex/edge-cloud-infra/billing"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

const BillingOrgGroup = "BillingOrg"

func init() {
	cmds := []*ApiCommand{&ApiCommand{
		Name:         "CreateBillingOrg",
		Use:          "create",
		Short:        "Set up a BillingOrganization and validate inputs",
		RequiredArgs: "name type firstname lastname email",
		OptionalArgs: "address address2 city country state postalcode phone",
		ReqData:      &ormapi.BillingOrganization{},
		Comments:     CreateBillingOrgComments,
		Path:         "/auth/billingorg/create",
	}, &ApiCommand{
		Name:         "SetAccountInfo",
		Use:          "updateaccountinfo",
		Short:        "Commit a BillingOrganization after validating it with our payment platform",
		RequiredArgs: "orgname accountid",
		OptionalArgs: "subscriptionid",
		ReqData:      &ormapi.AccountInfo{},
		Comments:     ormapi.AccountInfoComments,
		Path:         "/auth/billingorg/updateaccount",
	}, &ApiCommand{
		Name:         "UpdateBillingOrg",
		Use:          "update",
		Short:        "Update a billing organization",
		RequiredArgs: "name",
		OptionalArgs: "firstname lastname email address city country state postalcode",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Path:         "/auth/billingorg/update",
	}, &ApiCommand{
		Name:         "AddBillingOrgChild",
		Use:          "addchild",
		Short:        "Add an organization as a child of a billing organization",
		RequiredArgs: "name child",
		AliasArgs:    "child=children",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Path:         "/auth/billingorg/addchild",
	}, &ApiCommand{
		Name:         "RemoveBillingOrgChild",
		Use:          "removechild",
		Short:        "Remove an organization from a billing organization",
		RequiredArgs: "name child",
		AliasArgs:    "child=children",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Path:         "/auth/billingorg/removechild",
	}, &ApiCommand{
		Name:         "DeleteBillingOrg",
		Use:          "delete",
		Short:        "Delete a billing organization",
		RequiredArgs: "name",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Path:         "/auth/billingorg/delete",
	}, &ApiCommand{
		Name:      "ShowBillingOrg",
		Use:       "show",
		Short:     "Show billing organizations",
		Comments:  ormapi.BillingOrganizationComments,
		ReplyData: &[]ormapi.BillingOrganization{},
		Path:      "/auth/billingorg/show",
	}, &ApiCommand{
		Name:         "GetInvoice",
		Use:          "getinvoice",
		RequiredArgs: "name",
		OptionalArgs: "startdate enddate",
		ReqData:      &ormapi.InvoiceRequest{},
		Comments:     ormapi.InvoiceRequestComments,
		ReplyData:    &[]billing.InvoiceData{},
		Path:         "/auth/billingorg/invoice",
	}}
	AllApis.AddGroup(BillingOrgGroup, "Manage billing organizations", cmds)
}

var CreateBillingOrgComments = map[string]string{
	"name":       "name of the billingOrg",
	"type":       "type of the billingOrg",
	"firstname":  "First name",
	"lastname":   "Last name",
	"email":      "Email address",
	"address":    "Address line 1",
	"address2":   "Address line 2",
	"city":       "City",
	"country":    "Country",
	"state":      "State",
	"postalcode": "zip code",
	"phone":      "Phone number",
}
