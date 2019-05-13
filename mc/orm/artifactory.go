package orm

import (
	"context"
	"os"
	"strings"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/mobiledgex/edge-cloud/log"
)

const (
	GroupPrefix string = "mexdev-group-"
	RepoPrefix  string = "mexdev-repo-"
	PermPrefix  string = "mexdev-perm-"
)

func artifactoryConnected() bool {
	if artifactoryClient != nil {
		return true
	}
	return false
}

func artifactoryConnect() error {
	artifactoryApiKey := os.Getenv("artifactory_apikey")
	if artifactoryApiKey == "" {
		log.InfoLog("Note: No 'artifactory_apikey' env var found")
		// return success, no point re-connecting to artifactory
		return nil
	}
	tp := transport.ApiKeyAuth{
		ApiKey: artifactoryApiKey,
	}
	var err error
	artifactoryClient, err = artifactory.NewClient(
		serverConfig.ArtifactoryAddr,
		tp.Client())
	if err != nil {
		log.InfoLog("Note: Failed to connect to artifactory", "addr",
			serverConfig.ArtifactoryAddr, "err", err)
	}
	return err
}

func getGroupName(orgName string) string {
	return GroupPrefix + orgName
}

func getRepoName(orgName string) string {
	return RepoPrefix + orgName
}

func getPermTargetName(orgName string) string {
	return PermPrefix + orgName
}

func artifactoryListGroups() (map[string]bool, error) {
	groups, _, err := artifactoryClient.V1.Security.ListGroups(context.Background())
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]bool)
	for _, group := range *groups {
		groupName := *group.Name
		if strings.HasPrefix(groupName, GroupPrefix) {
			tmp[groupName] = true
		}
	}
	return tmp, nil
}

func artifactoryCreateGroup(orgName string) error {
	groupName := getGroupName(orgName)
	group := v1.Group{
		Name:            artifactory.String(groupName),
		Realm:           artifactory.String("ldap"),
		RealmAttributes: artifactory.String("ldapGroupName=" + orgName + ";groupsStrategy=STATIC;groupDn=cn=" + orgName + ",ou=orgs"),
	}
	_, err := artifactoryClient.V1.Security.CreateOrReplaceGroup(context.Background(), groupName, &group)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory create group",
			"group", groupName, "err", err)
	}
	return err
}

func artifactoryDeleteGroup(orgName string) error {
	groupName := getGroupName(orgName)
	_, _, err := artifactoryClient.V1.Security.DeleteGroup(context.Background(), groupName)
	if err != nil {
		if strings.Contains(err.Error(), "Status:404") {
			log.DebugLog(log.DebugLevelApi, "artifactory delete group",
				"group", groupName, "err", "group does not exists")
			return nil
		}
		log.DebugLog(log.DebugLevelApi, "artifactory delete group",
			"group", groupName, "err", err)
	}
	return err
}

func artifactoryListRepos() (map[string]bool, error) {
	repos, _, err := artifactoryClient.V1.Repositories.ListRepositories(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]bool)
	for _, repo := range *repos {
		repoName := *repo.Key
		if strings.HasPrefix(repoName, RepoPrefix) {
			tmp[repoName] = true
		}
	}
	return tmp, nil
}

func artifactoryCreateRepo(orgName string) error {
	repoName := getRepoName(orgName)
	repo := v1.LocalRepository{
		Key:             artifactory.String(repoName),
		RClass:          artifactory.String("local"),
		PackageType:     artifactory.String("generic"),
		HandleSnapshots: artifactory.Bool(false),
	}

	_, err := artifactoryClient.V1.Repositories.CreateLocal(context.Background(), &repo)
	if err != nil {
		if strings.Contains(err.Error(), "key already exists") {
			log.DebugLog(log.DebugLevelApi, "artifactory create repository",
				"repository", repoName, "err", "already exists")
			return nil
		}
		log.DebugLog(log.DebugLevelApi, "artifactory create repository",
			"repository", repoName, "err", err)
	}
	return err
}

func artifactoryDeleteRepo(orgName string) error {
	repoName := getRepoName(orgName)
	_, err := artifactoryClient.V1.Repositories.DeleteLocal(context.Background(), repoName)
	if err != nil {
		if strings.Contains(err.Error(), "Status:404") {
			log.DebugLog(log.DebugLevelApi, "artifactory delete repository",
				"repository", repoName, "err", "repository does not exists")
			return nil
		}
		log.DebugLog(log.DebugLevelApi, "artifactory delete repository",
			"repository", repoName, "err", err)
	}
	return err
}

func artifactoryListPerms() (map[string]bool, error) {
	perms, _, err := artifactoryClient.V1.Security.ListPermissionTargets(context.Background())
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]bool)
	for _, perm := range perms {
		permName := *perm.Name
		if strings.HasPrefix(permName, PermPrefix) {
			tmp[permName] = true
		}
	}
	return tmp, nil
}

func artifactoryCreateRepoPerms(orgName string) error {
	groupName := getGroupName(orgName)
	repoName := getRepoName(orgName)
	permTargetName := getPermTargetName(orgName)
	// create permission target
	permTargets := v1.PermissionTargets{
		Name:         artifactory.String(permTargetName),
		Repositories: &[]string{repoName},
		Principals: &v1.Principals{
			Groups: &map[string][]string{
				groupName: []string{"m", "w", "n", "r"},
			},
		},
	}
	_, err := artifactoryClient.V1.Security.CreateOrReplacePermissionTargets(context.Background(), permTargetName, &permTargets)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory create repo perms",
			"permission target", permTargetName, "repository", repoName, "group", groupName, "err", err)
	}
	return err
}

func artifactoryDeleteRepoPerms(orgName string) error {
	permTargetName := getPermTargetName(orgName)
	_, _, err := artifactoryClient.V1.Security.DeletePermissionTargets(context.Background(), permTargetName)
	if err != nil {
		if strings.Contains(err.Error(), "Status:404") {
			log.DebugLog(log.DebugLevelApi, "artifactory delete repo perms",
				"repo perms", permTargetName, "err", "repo perms does not exists")
			return nil
		}
		log.DebugLog(log.DebugLevelApi, "artifactory delete repo perms",
			"permission target", permTargetName, "err", err)
	}
	return err
}

func artifactoryCreateGroupObjects(orgName string) {
	if !artifactoryConnected() {
		return
	}
	err := artifactoryCreateGroup(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
	err = artifactoryCreateRepo(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
	err = artifactoryCreateRepoPerms(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
}

func artifactoryDeleteGroupObjects(orgName string) {
	if !artifactoryConnected() {
		return
	}
	err := artifactoryDeleteGroup(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
	err = artifactoryDeleteRepo(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
	err = artifactoryDeleteRepoPerms(orgName)
	if err != nil {
		artifactorySync.NeedsSync()
		return
	}
}
