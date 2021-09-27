package common

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
)

const (
	// Federation Types
	TypeSelf    = "self"
	TypePartner = "partner"

	// Partner federation member roles
	RoleAccessPartnerZones    = "access" // Self member can access partner member's zones
	RoleShareZonesWithPartner = "share"  // Self member will share its zones with partner member

	// Delimiter
	Delimiter = "|"
)

func IsValidFederationType(fedType string) error {
	if fedType != TypeSelf && fedType != TypePartner {
		return fmt.Errorf("Invalid federation type. Valid types are %s, %s", TypeSelf, TypePartner)
	}
	return nil
}

func AddToDelimitedList(delimitedList *string, addVal string) error {
	if delimitedList == nil {
		return fmt.Errorf("Invalid input delimited list")
	}
	if *delimitedList == "" {
		*delimitedList = addVal
		return nil
	}
	vals := strings.Split(*delimitedList, Delimiter)
	for _, val := range vals {
		if val == addVal {
			// value already present
			return nil
		}
	}
	vals = append(vals, addVal)
	*delimitedList = strings.Join(vals, Delimiter)
	return nil
}

func RemoveFromDelimitedList(delimitedList *string, rmVal string) error {
	if delimitedList == nil {
		return fmt.Errorf("Invalid input delimited list")
	}
	vals := strings.Split(*delimitedList, Delimiter)
	for ii, val := range vals {
		if val == rmVal {
			vals = append(vals[:ii], vals[ii+1:]...)
			break
		}
	}
	*delimitedList = strings.Join(vals, Delimiter)
	return nil
}

func GetValuesFromDelimitedList(delimitedList string) map[string]struct{} {
	vals := strings.Split(delimitedList, Delimiter)
	valMap := make(map[string]struct{})
	for _, v := range vals {
		valMap[v] = struct{}{}
	}
	return valMap
}

func ValueExistsInDelimitedList(delimitedList, val string) bool {
	valMap := GetValuesFromDelimitedList(delimitedList)
	_, matchRes := valMap[val]
	return matchRes
}

func AddOrUpdatePartnerFederatorRole(loggedDb *gorm.DB, selfFed, partnerFed *ormapi.Federator, role string) error {
	// Store partner federation role
	partnerRoleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerRole := ormapi.FederatorRole{}
	res := loggedDb.Where(&partnerRoleLookup).First(&partnerRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if res.RecordNotFound() {
		partnerRole = ormapi.FederatorRole{
			SelfFederationId:    selfFed.FederationId,
			PartnerFederationId: partnerFed.FederationId,
			Role:                role,
		}
		if err := loggedDb.Create(&partnerRole).Error; err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
				return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s",
					partnerFed.OperatorId, partnerFed.CountryCode)
			}
			return ormutil.DbErr(err)
		}
	} else {
		err := AddToDelimitedList(&partnerRole.Role, role)
		if err != nil {
			return err
		}
		err = loggedDb.Save(&partnerRole).Error
		if err != nil {
			return ormutil.DbErr(err)
		}
	}
	return nil
}

func FederationRoleExists(loggedDb *gorm.DB, selfFed, partnerFed *ormapi.Federator, role string) (bool, error) {
	partnerRoleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerRole := ormapi.FederatorRole{}
	res := loggedDb.Where(&partnerRoleLookup).First(&partnerRole)
	if !res.RecordNotFound() && res.Error != nil {
		return false, ormutil.DbErr(res.Error)
	}
	if res.RecordNotFound() || !ValueExistsInDelimitedList(partnerRole.Role, role) {
		return false, nil
	}
	return true, nil
}
