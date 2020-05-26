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

func ChefNodeRunStatus(ctx context.Context, client *chef.Client, nodeName string, startTime time.Time) ([]ChefStatusInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch chef node's run status", "node name", nodeName)
	nodeInfo, err := client.Nodes.Get(nodeName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get chef node info: %s, %v", nodeName, err)
	}
	if _, ok := nodeInfo.NormalAttributes["runstatus"]; !ok {
		//return nil, fmt.Errorf("unable to find runstatus attributes")
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find runstatus attributes")
		return nil, nil
	}
	runStatusAttr := nodeInfo.NormalAttributes["runstatus"]
	var runStatus ChefRunStatus

	config := mapstructure.DecoderConfig{
		DecodeHook: stringToDateTimeHook,
		Result:     &runStatus,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new decoder: %v", err)
	}

	err = decoder.Decode(runStatusAttr)
	if err != nil {
		return nil, fmt.Errorf("unable to decode runstatus attributes: %s, error: %v", runStatusAttr, err)
	}

	chefStartTime, err := time.Parse(ChefTimeLayout, runStatus.Start)
	if err != nil {
		//return nil, fmt.Errorf("unable to parse runstatus start time: %s, %v", runStatus.Start, err)
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to parse runstatus start time", "start time", runStatus.Start, "err", err)
	}

	var statusInfo []ChefStatusInfo

	if chefStartTime.Before(startTime) {
		// Ignore status info as it is of old run
		log.SpanLog(ctx, log.DebugLevelInfra, "ASHISH: TIME CHECK", "runstatus", runStatusAttr)
		log.SpanLog(ctx, log.DebugLevelInfra, "skipping chef node status info as it is of old run",
			"node name", nodeName, "runstatus time", chefStartTime, "start time", startTime)
		//return statusInfo, nil
	}

	failed := false
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

func ChefNodeCreate(ctx context.Context, client *chef.Client, nodeName, roleName string, attributes map[string]interface{}) error {
	nodeObj := chef.Node{
		Name:        nodeName,
		Environment: "_default",
		ChefType:    "node",
		JsonClass:   "Chef::Node",
		RunList: []string{
			"role[base]",
			roleName,
		},
		NormalAttributes: attributes,
	}
	_, err := client.Nodes.Post(nodeObj)
	if err != nil {
		return fmt.Errorf("failed to create node %s: %v", nodeName, err)
	}

	acl := chef.NewACL("update", []string{nodeName}, []string{"clients", "admins", "users"})
	err = client.ACLs.Put("nodes", nodeName, "update", acl)
	if err != nil {
		return fmt.Errorf("failed to setup update permissions for node %s: %v", nodeName, err)
	}
	return nil
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
