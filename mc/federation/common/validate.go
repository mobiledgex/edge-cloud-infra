package common

import (
	"fmt"
	"strings"

	valid "github.com/asaskevich/govalidator"
)

func ValidateCountryCode(countryCode string) error {
	if countryCode == "" {
		return fmt.Errorf("Missing country code")
	}
	if valid.IsISO3166Alpha2(countryCode) {
		return nil
	}
	return fmt.Errorf("Invalid country code %q. It must be a valid ISO 3166-1 Alpha-2 code for the country", countryCode)
}

func ValidateGeoLocation(geoLoc string) error {
	if geoLoc == "" {
		return fmt.Errorf("Missing geo location")
	}
	loc := strings.Split(geoLoc, ",")
	if len(loc) != 2 {
		return fmt.Errorf("Invalid geo location %q. Valid format: <LatInDecimal,LongInDecimal>", geoLoc)
	}
	lat, long := strings.TrimSpace(loc[0]), strings.TrimSpace(loc[1])
	if !valid.IsLatitude(lat) {
		return fmt.Errorf("Invalid latitude specified in geo location %q", lat)
	}
	if !valid.IsLongitude(long) {
		return fmt.Errorf("Invalid longitude specified in geo location %q", long)
	}
	return nil
}
