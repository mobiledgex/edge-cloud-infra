package e2esetup

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/yaml/v2"
)

const (
	ListSlackMessagesApi      = "/slack/listmessages"
	DeleteAllSlackMessagesApi = "/slack/deleteall"
	SlackWebhookApi           = "/slack/webhook"
)

// The structure below is a a partial slack webhook message
// that alertmanager sends to the slack endpoint
type TestSlackMsg struct {
	Channel     string `json:"channel"`
	Username    string `json:"username"`
	Attachments []struct {
		Title     string `json:"title"`
		TitleLink string `json:"title_link"`
		Text      string `json:"text"`
		Fallback  string `json:"fallback"`
	} `json:"attachments"`
}

func GetHttpServer(name string) *intprocess.HttpServer {
	if name == "" {
		return Deployment.HttpServers[0]
	}
	for _, server := range Deployment.HttpServers {
		if server.Name == name {
			return server
		}
	}
	log.Fatalf("Error: could not find http server process: %s\n", name)
	return nil
}

// get api
func RunSlackAPI(api, apiFile, outputDir string) error {
	switch api {
	case "check":
		// get default
		proc := GetHttpServer("")
		apiUrl := fmt.Sprintf("0.0.0.0:%d%s", proc.Port, ListSlackMessagesApi)
		cmd := exec.Command("curl", "-s", "-S", apiUrl)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error running show slack messages API on port %d, err: %v\n", proc.Port, err)
			return err
		}
		// unmarshal and marshal back to get just the fields we want
		msgs := []TestSlackMsg{}
		err = json.Unmarshal(out, &msgs)
		if err != nil {
			log.Printf("slack message unmarshal error: %v\n", err)
			return err
		}
		// marshal back
		ymlOut, err := yaml.Marshal(&msgs)
		if err != nil {
			log.Printf("slack marshal into yaml error: %v\n", err)
			return err
		}
		truncate := true
		util.PrintToFile("show-commands.yml", outputDir, string(ymlOut), truncate)
	case "deleteall":
		// get default
		proc := GetHttpServer("")
		apiUrl := fmt.Sprintf("0.0.0.0:%d%s", proc.Port, DeleteAllSlackMessagesApi)
		cmd := exec.Command("curl", "-s", "-S", "-X", "DELETE", apiUrl)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("err: %v\n", err)
			return err
		}
	default:
		return fmt.Errorf("Unknown action for mock slack subsystem")
	}
	return nil
}
