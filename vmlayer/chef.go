package vmlayer

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chef/chef"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
)

type ChefAuthKey struct {
	ApiKey        string `json:"apikey"`
	ValidationKey string `json:"validationkey"`
}

func GetChefAuthKeys(ctx context.Context, vaultConfig *vault.Config, chefServerPath string) (*ChefAuthKey, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch chef auth keys", "chefServerPath", chefServerPath)
	urlObj, err := util.ImagePathParse(chefServerPath)
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(urlObj.Host, ":")
	if len(hostname) < 1 {
		return nil, fmt.Errorf("chef server path is empty")
	}
	vaultPath := cloudcommon.GetVaultRegistryPath(hostname[0])
	auth := &ChefAuthKey{}
	err = vault.GetData(vaultConfig, vaultPath, 0, auth)
	if err != nil {
		return nil, fmt.Errorf("Unable to find chef auth keys from vault path %s, %v", vaultPath, err)
	}
	if auth.ApiKey == "" {
		return nil, fmt.Errorf("Unable to find chef API key")
	}
	if auth.ValidationKey == "" {
		return nil, fmt.Errorf("Unable to find chef validation key")
	}
	return auth, nil
}

func GetChefClient(ctx context.Context, apiKey, chefServerPath string) (*chef.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "get chef client", "chefServerPath", chefServerPath)
	if !strings.HasSuffix(chefServerPath, "/") {
		chefServerPath = chefServerPath + "/"
	}
	client, err := chef.NewClient(&chef.Config{
		Name:    "chefadmin",
		Key:     apiKey,
		SkipSSL: true,
		BaseURL: chefServerPath,
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to setup chef client: %v", err)
	}
	return client, nil
}

func ChefRoleExists(ctx context.Context, client *chef.Client, roleName string) (bool, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "check if chef role exists", "role name", roleName)
	roleList, err := client.Roles.List()
	if err != nil {
		return false, fmt.Errorf("Unable to get chef roles: %v", err)
	}
	for k, _ := range *roleList {
		if k == roleName {
			return true, nil
		}
	}
	return false, nil
}

func ChefClientExists(ctx context.Context, client *chef.Client, clientName string) (bool, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "check if chef client exists", "client name", clientName)
	clientList, err := client.Clients.List()
	if err != nil {
		return false, fmt.Errorf("Unable to get chef clients: %v", err)
	}
	for k, _ := range clientList {
		if k == clientName {
			return true, nil
		}
	}
	return false, nil
}

func ChefClientDelete(ctx context.Context, client *chef.Client, clientName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete chef client", "client name", clientName)
	clientList, err := client.Clients.List()
	if err != nil {
		return fmt.Errorf("Unable to get chef clients: %v", err)
	}
	for k, _ := range clientList {
		if k == clientName {
			return client.Clients.Delete(k)
		}
	}
	return nil
}
