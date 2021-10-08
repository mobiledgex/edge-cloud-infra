package common

import (
	"fmt"
)

func FederatorStr(operatorId, countryCode string) string {
	return fmt.Sprintf("OperatorID: %q/CountryCode: %q", operatorId, countryCode)
}
