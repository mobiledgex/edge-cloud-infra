package e2esetup

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

// go-cmp Options
var IgnoreAdminRole = cmpopts.AcyclicTransformer("removeAdminRole", func(roles []ormapi.Role) []ormapi.Role {
	// remove automatically created admin role
	newroles := make([]ormapi.Role, 0)
	for _, role := range roles {
		if role.Username == "mexadmin" {
			continue
		}
		newroles = append(newroles, role)
	}
	sort.Slice(newroles, func(i, j int) bool {
		if newroles[i].Org < newroles[j].Org {
			return true
		}
		if newroles[i].Org > newroles[j].Org {
			return false
		}
		if newroles[i].Username < newroles[j].Username {
			return true
		}
		if newroles[i].Username > newroles[j].Username {
			return false
		}
		return newroles[i].Role < newroles[j].Role
	})
	return newroles
})

var IgnoreAdminUser = cmpopts.AcyclicTransformer("removeAdminUser", func(users []ormapi.User) []ormapi.User {
	// remove automatically created super user
	newusers := make([]ormapi.User, 0)
	for _, user := range users {
		if user.Name == "mexadmin" {
			continue
		}
		newusers = append(newusers, user)
	}
	return newusers
})

var IgnoreTaskStatusMessages = cmpopts.AcyclicTransformer("ignoreTaskStatus", func(ar ormapi.AuditResponse) ormapi.AuditResponse {
	if !strings.Contains(ar.OperationName, "/data/create") &&
		!strings.Contains(ar.OperationName, "/ctrl/CreateClusterInst") &&
		!strings.Contains(ar.OperationName, "/ctrl/CreateAppInst") &&
		!strings.Contains(ar.OperationName, "/ctrl/CreateCloudlet") {
		return ar
	}
	// Due to the way task/status updates are sent back from CRM to controller,
	// it's possible on the CRM side that quick sequential updates may
	// overwrite earlier updates, because the ClusterInstInfo/AppInstInfo/
	// CloudletInfo structs are kept in a cache on the notify sender thread.
	// This means it's possible (especially for the fake platform) that
	// some task+status updates may be lost. This means the audit response
	// output is non-deterministic, based on timing, which means we can't
	// just string compare it for the e2e tests. Making it deterministic is
	// probably not worth the effort, so instead here we remove the
	// updates that we know may not make it.
	resps := strings.Split(ar.Response, "\n")
	newResps := []string{}
	for _, resp := range resps {
		// substrings are based on the updateCallbacks from fake.go
		// (the fake platform)
		if strings.Contains(resp, "First Create Task") ||
			strings.Contains(resp, "Second Create Task") ||
			strings.Contains(resp, "Creating Cloudlet") ||
			strings.Contains(resp, "Creating App Inst") ||
			strings.Contains(resp, "Starting CRMServer") ||
			strings.Contains(resp, "fake appInst updated") {
			continue
		}
		newResps = append(newResps, resp)
	}
	ar.Response = strings.Join(newResps, "\n")
	return ar
})

func CmpSortOrgs(a ormapi.Organization, b ormapi.Organization) bool {
	return a.Name < b.Name
}

//compares two yaml files for equivalence
//TODO need to handle different types of interfaces besides appdata, currently using
//that to sort
func CompareYamlFiles(firstYamlFile string, secondYamlFile string, fileType string) bool {
	var err1 error
	var err2 error
	var y1 interface{}
	var y2 interface{}
	copts := []cmp.Option{}

	if fileType == "mcdata" {
		var a1 ormapi.AllData
		var a2 ormapi.AllData

		err1 = util.ReadYamlFile(firstYamlFile, &a1)
		err2 = util.ReadYamlFile(secondYamlFile, &a2)

		copts = []cmp.Option{
			cmpopts.IgnoreTypes(time.Time{}, dmeproto.Timestamp{}),
			IgnoreAdminRole,
		}
		copts = append(copts, edgeproto.IgnoreTaggedFields("nocmp")...)
		copts = append(copts, edgeproto.CmpSortSlices()...)
		copts = append(copts, cmpopts.SortSlices(CmpSortOrgs))

		y1 = a1
		y2 = a2
	} else if fileType == "mcusers" {
		// remove roles
		var a1 []ormapi.User
		var a2 []ormapi.User

		err1 = util.ReadYamlFile(firstYamlFile, &a1)
		err2 = util.ReadYamlFile(secondYamlFile, &a2)

		copts = []cmp.Option{
			cmpopts.IgnoreTypes(time.Time{}),
			IgnoreAdminUser,
		}
		y1 = a1
		y2 = a2
	} else if fileType == "mcaudit" {
		var a1 []ormapi.AuditResponse
		var a2 []ormapi.AuditResponse

		err1 = util.ReadYamlFile(firstYamlFile, &a1)
		err2 = util.ReadYamlFile(secondYamlFile, &a2)

		copts = []cmp.Option{
			cmpopts.IgnoreFields(ormapi.AuditResponse{}, "StartTime", "Duration", "TraceID"),
			IgnoreTaskStatusMessages,
		}
		y1 = a1
		y2 = a2
	} else if fileType == "mcmetrics" {
		var a1 []ormapi.MetricsCompare
		var a2 []ormapi.MetricsCompare

		err1 = util.ReadYamlFile(firstYamlFile, &a1)
		err2 = util.ReadYamlFile(secondYamlFile, &a2)

		y1 = a1
		y2 = a2

	} else {
		return util.CompareYamlFiles(firstYamlFile,
			secondYamlFile, fileType)
	}

	util.PrintStepBanner("running compareYamlFiles")
	log.Printf("Comparing yamls: %v  %v\n", firstYamlFile, secondYamlFile)

	if err1 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", firstYamlFile, err1)
		return false
	}
	if err2 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", secondYamlFile, err2)
		return false
	}

	if !cmp.Equal(y1, y2, copts...) {
		log.Println("Comparison fail")
		log.Printf(cmp.Diff(y1, y2, copts...))
		return false
	}
	log.Println("Comparison success")
	return true
}
