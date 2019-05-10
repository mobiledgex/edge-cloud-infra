package orm

import (
	"context"
	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/mobiledgex/edge-cloud/log"
)

func artifactoryConnected() bool {
	if artifactoryClient != nil {
		return true
	}
	return false
}

func getGroupName(orgName string) string {
	return "mexdev-group-" + orgName
}

func getRepoName(orgName string) string {
	return "mexdev-repo-" + orgName
}

func getPermTargetName(orgName string) string {
	return "mexdev-perm-" + orgName
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
		log.DebugLog(log.DebugLevelApi, "artifactory delete group",
			"group", groupName, "err", err)
	}
	return err
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
		log.DebugLog(log.DebugLevelApi, "artifactory create repository",
			"repository", repoName, "err", err)
	}
	return err
}

func artifactoryDeleteRepo(orgName string) error {
	repoName := getRepoName(orgName)
	_, err := artifactoryClient.V1.Repositories.DeleteLocal(context.Background(), repoName)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory delete repository",
			"repository", repoName, "err", err)
	}
	return err
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
		return
	}
	err = artifactoryCreateRepo(orgName)
	if err != nil {
		return
	}
	err = artifactoryCreateRepoPerms(orgName)
	if err != nil {
		return
	}
}

func artifactoryDeleteGroupObjects(orgName string) {
	if !artifactoryConnected() {
		return
	}
	err := artifactoryDeleteGroup(orgName)
	if err != nil {
		return
	}
	err = artifactoryDeleteRepo(orgName)
	if err != nil {
		return
	}
	err = artifactoryDeleteRepoPerms(orgName)
	if err != nil {
		return
	}
}

func artifactoryCreateAllGroupObjects() {
	if !artifactoryConnected() {
		return
	}
	log.DebugLog(log.DebugLevelApi, "artifactory create all group objects")
	groupings := enforcer.GetGroupingPolicy()
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil || role.Org == "" {
			continue
		}
		artifactoryCreateGroupObjects(role.Org)
	}
}
