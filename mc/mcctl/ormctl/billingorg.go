package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetBillingOrgCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "create",
		Short:        "Create a billing organization to handle billing",
		RequiredArgs: "name type firstname lastname email",
		OptionalArgs: "address address2 city country state postalcode phone paymenttype ccfirstname cclastname ccnumber ccexpmonth ccexpyear children",
		AliasArgs:    strings.Join(CreateBillingOrgAliasArgs, " "),
		ReqData:      &ormapi.CreateBillingOrganization{},
		Comments:     CreateBillingOrgComments,
		Run:          runRest("/auth/billingorg/create"),
	}, &cli.Command{
		Use:          "update",
		Short:        "Update a billing organization",
		RequiredArgs: "name",
		OptionalArgs: "firstname lastname email address city country state postalcode",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Run:          runRest("/auth/billingorg/update"),
	}, &cli.Command{
		Use:          "addchild",
		Short:        "Add an organization as a child of a billing organization",
		RequiredArgs: "name child",
		AliasArgs:    "child=children",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Run:          runRest("/auth/billingorg/addchild"),
	}, &cli.Command{
		Use:          "removechild",
		Short:        "Remove an organization from a billing organization",
		RequiredArgs: "name child",
		AliasArgs:    "child=children",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Run:          runRest("/auth/billingorg/removechild"),
	}, &cli.Command{
		Use:          "delete",
		Short:        "Delete a billing organization",
		RequiredArgs: "name",
		Comments:     ormapi.BillingOrganizationComments,
		ReqData:      &ormapi.BillingOrganization{},
		Run:          runRest("/auth/billingorg/delete"),
	}, &cli.Command{
		Use:       "show",
		Short:     "Show billing organizations",
		Comments:  ormapi.BillingOrganizationComments,
		ReplyData: &[]ormapi.BillingOrganization{},
		Run:       runRest("/auth/billingorg/show"),
	}}
	return cli.GenGroup("billingorg", "Manage billing organizations", cmds)
}

var CreateBillingOrgAliasArgs = []string{
	"paymenttype=payment.paymenttype",
	"ccfirstname=payment.creditcard.firstname",
	"cclastname=payment.creditcard.lastname",
	"ccnumber=payment.creditcard.cardnumber",
	"ccexpmonth=payment.creditcard.expirationmonth",
	"ccexpyear=payment.creditcard.expirationyear",
}

var CreateBillingOrgComments = map[string]string{
	"name":        "name of the billingOrg",
	"type":        "type of the billingOrg",
	"firstname":   "First name",
	"lastname":    "Last name",
	"email":       "Email address",
	"address":     "Address line 1",
	"address2":    "Address line 2",
	"city":        "City",
	"country":     "Country",
	"state":       "State",
	"postalcode":  "zip code",
	"phone":       "Phone number",
	"paymenttype": "payment type, currently supported methods are: `credit_card`",
	"ccfirstname": "First Name as appears on the credit card",
	"cclastname":  "Last Name as appears on the credit card",
	"ccnumber":    "Credit card number",
	"ccexpmonth":  "Credit card expiration month (mm)",
	"ccexpyear":   "Credit card expiration year (yyyy)",
}
