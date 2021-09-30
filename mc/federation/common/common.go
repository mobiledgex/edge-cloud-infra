package common

import (
	"fmt"
)

type FederatorRole string

const (
	Delimiter = "|"

	TypeSelf    = "self"
	TypePartner = "partner"
)

var (
	RoleAny FederatorRole = "*"
	// Partner federator can share zones with self federator
	RoleShareZonesWithSelf FederatorRole = "share-zones-with-self"
	// Partner federator can access zones of self federator
	RoleAccessToSelfZones FederatorRole = "access-to-self-zones"
)

func FederatorStr(operatorId, countryCode string) string {
	return fmt.Sprintf("OperatorID: %q/CountryCode: %q", operatorId, countryCode)
}
