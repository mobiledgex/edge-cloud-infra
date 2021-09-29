package common

import (
	"fmt"
)

const (
	Delimiter = "|"
)

func FederatorStr(operatorId, countryCode string) string {
	return fmt.Sprintf("OperatorID: %q/CountryCode: %q", operatorId, countryCode)
}
