package e2esetup

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/yaml/v2"
)

const (
	ListPagerDutyMessagesApi    = "/pagerduty/listevents"
	DeleteAllPagerDutyEventsApi = "/pagerduty/deleteall"
	PagerDutyApi                = "/pagerduty/event"
)

// The structure below is a a partial pagerduty event
// that alertmanager sends to the pagerduty api
type TestPagerDutyEvent struct {
	RoutingKey  string `json:"routing_key"`
	EventAction string `json:"event_action"`
	Payload     struct {
		Summary       string `json:"summary"`
		Source        string `json:"source"`
		Severity      string `json:"severity"`
		CustomDetails struct {
			Firing      string `json:"firing"`
			NumFiring   string `json:"num_firing"`
			NumResolved string `json:"num_resolved"`
			Resolved    string `json:"resolved"`
		} `json:"custom_details"`
	} `json:"payload"`
	Client    string `json:"client"`
	ClientURL string `json:"client_url"`
}

// get api
func RunPagerDutyAPI(api, apiFile, outputDir string) error {
	servers := make([]E2eServerName, 0)
	if apiFile != "" {
		err := util.ReadYamlFile(apiFile, &servers)
		if err != nil {
			log.Printf("Unable to read api file: %s [%s]\n", apiFile, err.Error())
			return err
		}
	} else {
		servers = append(servers, E2eServerName{Name: ""})
	}

	switch api {
	case "check":
		for ii, sName := range servers {
			proc := GetHttpServer(sName.Name)
			apiUrl := fmt.Sprintf("0.0.0.0:%d%s", proc.Port, ListPagerDutyMessagesApi)
			cmd := exec.Command("curl", "-s", "-S", apiUrl)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error running show pagerduty events API on port %d, err: %v\n", proc.Port, err)
				return err
			}
			// unmarshal and marshal back to get just the fields we want
			msgs := []TestPagerDutyEvent{}
			err = json.Unmarshal(out, &msgs)
			if err != nil {
				log.Printf("pagerduty event unmarshal error: %v\n", err)
				return err
			}
			// marshal back
			ymlOut, err := yaml.Marshal(&msgs)
			if err != nil {
				log.Printf("pagerduty marshal into yaml error: %v\n", err)
				return err
			}
			if ii == 0 {
				util.PrintToFile("show-commands.yml", outputDir, string(ymlOut), true)
			} else {
				util.PrintToFile("show-commands.yml", outputDir, string(ymlOut), false)
			}
		}
	case "deleteall":
		for _, sName := range servers {
			proc := GetHttpServer(sName.Name)
			apiUrl := fmt.Sprintf("0.0.0.0:%d%s", proc.Port, DeleteAllPagerDutyEventsApi)
			cmd := exec.Command("curl", "-s", "-S", "-X", "DELETE", apiUrl)
			_, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("err: %v\n", err)
				return err
			}
		}
	default:
		return fmt.Errorf("Unknown action for mock pagerduty subsystem")
	}
	return nil
}
