package federation

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

func AddFederationRole(fedObj *ormapi.OperatorFederation, inRole string) error {
	if fedObj.Role == "" {
		fedObj.Role = inRole
		return nil
	}
	roles := strings.Split(fedObj.Role, ",")
	for _, role := range roles {
		if role == inRole {
			// role already present
			return nil
		}
	}
	roles = append(roles, inRole)
	fedObj.Role = strings.Join(roles, ",")
	return nil
}

func RemoveFederationRole(fedObj *ormapi.OperatorFederation, inRole string) error {
	roles := strings.Split(fedObj.Role, ",")
	for ii, role := range roles {
		if role == inRole {
			roles = append(roles[:ii], roles[ii+1:]...)
			break
		}
	}
	fedObj.Role = strings.Join(roles, ",")
	return nil
}

func FederationRoleExists(fedObj *ormapi.OperatorFederation, inRole string) bool {
	roles := strings.Split(fedObj.Role, ",")
	roleMap := make(map[string]struct{})
	for _, role := range roles {
		roleMap[role] = struct{}{}
	}
	_, matchRes := roleMap[inRole]
	return matchRes
}
