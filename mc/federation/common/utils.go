package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	valid "github.com/asaskevich/govalidator"
)

func FederatorStr(operatorId, countryCode string) string {
	return fmt.Sprintf("OperatorID: %q/CountryCode: %q", operatorId, countryCode)
}

func ParseGeoLocation(geoLoc string) (float64, float64, error) {
	var lat float64
	var long float64
	if geoLoc == "" {
		return lat, long, fmt.Errorf("Missing geo location")
	}
	loc := strings.Split(geoLoc, ",")
	if len(loc) != 2 {
		return lat, long, fmt.Errorf("Invalid geo location %q. Valid format: <LatInDecimal,LongInDecimal>", geoLoc)
	}
	latStr, longStr := strings.TrimSpace(loc[0]), strings.TrimSpace(loc[1])
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return lat, long, err
	}

	long, err = strconv.ParseFloat(longStr, 64)
	if err != nil {
		return lat, long, err
	}

	return lat, long, nil
}

// NOTE: '.' should be not be allowed as it is used for parse federation ID
//       from zone ID
var zoneIdMatch = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$")

func ValidateZoneId(zoneId string) error {
	if !zoneIdMatch.MatchString(zoneId) {
		return fmt.Errorf("Invalid zone ID %q, valid format is %s", zoneId, zoneIdMatch)
	}
	return nil
}

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
