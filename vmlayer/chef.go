package vmlayer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-chef/chef"
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type ChefAuthKey struct {
	ApiKey        string `json:"apikey"`
	ValidationKey string `json:"validationkey"`
}

type ChefResource struct {
	CookbookName string   `mapstructure:"cookbook_name"`
	RecipeName   string   `mapstructure:"recipe_name"`
	Action       []string `mapstructure:"action"`
	Resource     string   `mapstructure:"resource"`
	ResourceType string   `mapstructure:"resource_type"`
	Updated      bool     `mapstructure:"updated"`
}

type ChefRunStatus struct {
	Status    string         `mapstructure:"status"`
	Start     string         `mapstructure:"start"`
	End       string         `mapstructure:"end"`
	Backtrace []string       `mapstructure:"backtrace"`
	Exception string         `mapstructure:"exception"`
	Resources []ChefResource `mapstructure:"resources"`
}

type ChefStatusInfo struct {
	Message string
	Failed  bool
}

const (
	ResourceDockerRegistry  = "docker_registry"
	ResourceDockerImage     = "docker_image"
	ResourceDockerContainer = "docker_container"
	ChefTimeLayout          = "2006-01-02 15:04:05 +0000"
)

func GetChefAuthKeys(ctx context.Context, vaultConfig *vault.Config) (*ChefAuthKey, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch chef auth keys")
	vaultPath := "/secret/data/accounts/chef"
	auth := &ChefAuthKey{}
	err := vault.GetData(vaultConfig, vaultPath, 0, auth)
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
			err = client.Clients.Delete(k)
			if err != nil {
				return err
			}
			break
		}
	}
	nodeList, err := client.Nodes.List()
	if err != nil {
		return fmt.Errorf("unable to get chef nodes: %v", err)
	}
	for k, _ := range nodeList {
		if k == clientName {
			return client.Nodes.Delete(k)
		}
	}
	return nil
}

func stringToDateTimeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t == reflect.TypeOf(time.Time{}) && f == reflect.TypeOf("") {
		return time.Parse(ChefTimeLayout, data.(string))
	}
	return data, nil
}

func ChefClientRunStatus(ctx context.Context, client *chef.Client, clientName string, startTime time.Time) ([]ChefStatusInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch chef client's run status", "client name", clientName)
	nodeInfo, err := client.Nodes.Get(clientName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get chef node info: %s, %v", clientName, err)
	}
	if _, ok := nodeInfo.NormalAttributes["runstatus"]; !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find runstatus attributes")
		return nil, nil
	}
	runStatusAttr := nodeInfo.NormalAttributes["runstatus"]

	var runStatus ChefRunStatus
	err = mapstructure.Decode(runStatusAttr, &runStatus)
	if err != nil {
		return nil, fmt.Errorf("unable to decode runstatus attributes: %s, error: %v", runStatusAttr, err)
	}

	failed := false
	var statusInfo []ChefStatusInfo
	for ii, res := range runStatus.Resources {
		msg := ""
		switch res.ResourceType {
		case ResourceDockerRegistry:
			msg = "Log in to docker registry"
		case ResourceDockerImage:
			msg = "Fetch docker image to start cloudlet services"
		case ResourceDockerContainer:
			msg = fmt.Sprintf("Start %s", res.Resource)
		default:
			msg = res.Resource
		}

		if ii == len(statusInfo)-1 && runStatus.Exception != "" {
			msg = fmt.Sprintf("Failed to %s", msg)
			failed = true
			log.SpanLog(ctx, log.DebugLevelInfra, "failure message from chef node run status", "message", msg, "exception", runStatus.Exception)
		}

		if !failed && !res.Updated {
			log.SpanLog(ctx, log.DebugLevelInfra, "skipping chef node status as it is not executed", "message", msg)
			continue
		}

		statusInfo = append(statusInfo, ChefStatusInfo{
			Message: msg,
			Failed:  failed,
		})
	}
	return statusInfo, nil
}

func ChefClientCreate(ctx context.Context, client *chef.Client, clientName, roleName string, attributes map[string]interface{}) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "chef client create", "client name", clientName, "role name", roleName)
	clientObj := chef.ApiNewClient{
		Name:      clientName,
		Validator: false,
		Admin:     false,
		CreateKey: true,
	}
	out, err := client.Clients.Create(clientObj)
	if err != nil {
		return "", fmt.Errorf("failed to create client %s: %v", clientName, err)
	}
	clientKey := out.ChefKey.PrivateKey
	if clientKey == "" {
		return "", fmt.Errorf("unable to get private key of the client %s", clientName)
	}

	nodeObj := chef.Node{
		Name:        clientName,
		Environment: "_default",
		ChefType:    "node",
		JsonClass:   "Chef::Node",
		RunList: []string{
			"role[base]",
			roleName,
		},
		NormalAttributes: attributes,
	}
	_, err = client.Nodes.Post(nodeObj)
	if err != nil {
		return "", fmt.Errorf("failed to create node %s: %v", clientName, err)
	}

	acl := chef.NewACL("update", []string{clientName}, []string{"admins", "users"})
	err = client.ACLs.Put("nodes", clientName, "update", acl)
	if err != nil {
		return "", fmt.Errorf("unable to add update acl for node %s", clientName)
	}
	return clientKey, nil
}

func GetChefArgs(ctx context.Context, cmdArgs []string) map[string]string {
	chefArgs := make(map[string]string)
	ii := 0
	for ii < len(cmdArgs) {
		if !strings.HasPrefix(cmdArgs[ii], "-") {
			continue
		}
		argKey := strings.TrimLeft(cmdArgs[ii], "-")
		argVal := ""
		ii += 1
		if ii < len(cmdArgs) && !strings.HasPrefix(cmdArgs[ii], "-") {
			argVal = cmdArgs[ii]
			ii += 1
		}
		chefArgs[argKey] = argVal
	}
	return chefArgs
}
