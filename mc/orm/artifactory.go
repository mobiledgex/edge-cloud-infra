package orm

import (
	"context"
	"strings"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	rtf "github.com/mobiledgex/edge-cloud-infra/artifactory"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/log"
)

const ArtifactoryPrefix string = "mc-"

func getArtifactoryName(orgName string) string {
	return ArtifactoryPrefix + orgName
}

func artifactoryClient() (*artifactory.Artifactory, error) {
	artifactoryApiKey, err := rtf.GetArtifactoryApiKey()
	if err != nil {
		return nil, err
	}
	tp := transport.ApiKeyAuth{
		ApiKey: artifactoryApiKey,
	}
	client, err := artifactory.NewClient(serverConfig.ArtifactoryAddr, tp.Client())
	if err != nil {
		log.InfoLog("Note: Failed to connect to artifactory", "addr",
			serverConfig.ArtifactoryAddr, "err", err)
		return nil, err
	}
	return client, nil
}

func artifactoryListUsers() (map[string]struct{}, error) {
	client, err := artifactoryClient()
	if err != nil {
		return nil, err
	}
	users, _, err := client.V1.Security.ListUsers(context.Background())
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]struct{})
	for _, user := range *users {
		userName := *user.Name
		if *user.Realm == "ldap" && userName != "admin" {
			tmp[userName] = struct{}{}
		}
	}
	return tmp, nil
}

func artifactoryListUserGroups(userName string) (map[string]struct{}, error) {
	client, err := artifactoryClient()
	if err != nil {
		return nil, err
	}

	tmp := make(map[string]struct{})
	userInfo, _, err := client.V1.Security.GetUser(context.Background(), userName)
	if err == nil && userInfo.Groups != nil {
		for _, group := range *userInfo.Groups {
			tmp[group] = struct{}{}
		}
	}
	return tmp, nil
}

func artifactoryCreateUser(user *ormapi.User, groups *[]string) {
	client, err := artifactoryClient()
	userName := user.Name
	if err == nil {
		rtfUser := v1.User{
			Name:                     artifactory.String(userName),
			Email:                    artifactory.String(user.Email),
			ProfileUpdatable:         artifactory.Bool(false),
			Groups:                   groups,
			InternalPasswordDisabled: artifactory.Bool(true),
		}
		_, err = client.V1.Security.CreateOrReplaceUser(context.Background(), userName, &rtfUser)
	}
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory create user",
			"user", userName, "err", err)
		artifactorySync.NeedsSync()
		return
	}
}

func artifactoryDeleteUser(userName string) {
	client, err := artifactoryClient()
	if err == nil {
		_, _, err = client.V1.Security.DeleteUser(context.Background(), userName)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Status:404") {
			log.DebugLog(log.DebugLevelApi, "artifactory delete user",
				"user", userName, "err", "user does not exists")
			return
		}
		log.DebugLog(log.DebugLevelApi, "artifactory delete user",
			"user", userName, "err", err)
		artifactorySync.NeedsSync()
		return
	}
}

func artifactoryAddUserToGroup(role *ormapi.Role) {
	client, err := artifactoryClient()
	userName := role.Username
	orgName := getArtifactoryName(role.Org)
	log.DebugLog(log.DebugLevelApi, "artifactory add user to group",
		"user", userName, "group", orgName)
	if err == nil {
		var userInfo *v1.User
		userInfo, _, err = client.V1.Security.GetUser(context.Background(), userName)
		if err == nil {
			var groups []string
			if userInfo.Groups != nil {
				groups = *userInfo.Groups
			}
			groups = append(groups, orgName)
			rtfUser := v1.User{
				Name:   artifactory.String(userName),
				Groups: &groups,
			}
			_, err = client.V1.Security.UpdateUser(context.Background(), userName, &rtfUser)
		}
	}
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory add user to group",
			"user", userName, "group", orgName, "err", err)
		artifactorySync.NeedsSync()
		return
	}
}

func artifactoryRemoveUserFromGroup(role *ormapi.Role) {
	client, err := artifactoryClient()
	userName := role.Username
	orgName := getArtifactoryName(role.Org)
	log.DebugLog(log.DebugLevelApi, "artifactory remove user from group",
		"user", userName)
	if err == nil {
		var userInfo *v1.User
		userInfo, _, err = client.V1.Security.GetUser(context.Background(), userName)
		if err == nil && userInfo.Groups != nil {
			var groups []string
			for _, group := range *userInfo.Groups {
				if group != orgName {
					groups = append(groups, group)
				}
			}
			rtfUser := v1.User{
				Name:   artifactory.String(userName),
				Groups: &groups,
			}
			_, err = client.V1.Security.UpdateUser(context.Background(), userName, &rtfUser)
		}
	}
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory remove user from group",
			"user", userName, "err", err)
		artifactorySync.NeedsSync()
		return
	}
}

func artifactoryListGroups() (map[string]struct{}, error) {
	client, err := artifactoryClient()
	if err != nil {
		return nil, err
	}
	groups, _, err := client.V1.Security.ListGroups(context.Background())
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]struct{})
	for _, group := range *groups {
		groupName := *group.Name
		if strings.HasPrefix(groupName, ArtifactoryPrefix) {
			tmp[groupName] = struct{}{}
		}
	}
	return tmp, nil
}

func artifactoryCreateGroup(orgName string) error {
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	groupName := getArtifactoryName(orgName)
	group := v1.Group{
		Name:        artifactory.String(groupName),
		Description: artifactory.String("Group maintained by master-controller"),
	}
	_, err = client.V1.Security.CreateOrReplaceGroup(context.Background(), groupName, &group)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory create group",
			"group", groupName, "err", err)
	}
	return err
}

func artifactoryDeleteGroup(orgName string) error {
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	groupName := getArtifactoryName(orgName)
	_, _, err = client.V1.Security.DeleteGroup(context.Background(), groupName)
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

func artifactoryListRepos() (map[string]struct{}, error) {
	client, err := artifactoryClient()
	if err != nil {
		return nil, err
	}
	repos, _, err := client.V1.Repositories.ListRepositories(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]struct{})
	for _, repo := range *repos {
		repoName := *repo.Key
		if strings.HasPrefix(repoName, ArtifactoryPrefix) {
			tmp[repoName] = struct{}{}
		}
	}
	return tmp, nil
}

func artifactoryCreateRepo(orgName string) error {
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	repoName := getArtifactoryName(orgName)
	repo := v1.LocalRepository{
		Key:             artifactory.String(repoName),
		RClass:          artifactory.String("local"),
		PackageType:     artifactory.String("generic"),
		HandleSnapshots: artifactory.Bool(false),
	}

	_, err = client.V1.Repositories.CreateLocal(context.Background(), &repo)
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
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	repoName := getArtifactoryName(orgName)
	_, err = client.V1.Repositories.DeleteLocal(context.Background(), repoName)
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

func artifactoryListPerms() (map[string]struct{}, error) {
	client, err := artifactoryClient()
	if err != nil {
		return nil, err
	}
	perms, _, err := client.V1.Security.ListPermissionTargets(context.Background())
	if err != nil {
		return nil, err
	}
	tmp := make(map[string]struct{})
	for _, perm := range perms {
		permName := *perm.Name
		if strings.HasPrefix(permName, ArtifactoryPrefix) {
			tmp[permName] = struct{}{}
		}
	}
	return tmp, nil
}

func artifactoryCreateRepoPerms(orgName string) error {
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	groupName := getArtifactoryName(orgName)
	repoName := getArtifactoryName(orgName)
	permTargetName := getArtifactoryName(orgName)
	// create permission target
	permTargets := v1.PermissionTargets{
		Name:         artifactory.String(permTargetName),
		Repositories: &[]string{repoName},
		Principals: &v1.Principals{
			Groups: &map[string][]string{
				groupName: []string{"w", "d", "r"},
			},
		},
	}
	_, err = client.V1.Security.CreateOrReplacePermissionTargets(context.Background(), permTargetName, &permTargets)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "artifactory create repo perms",
			"permission target", permTargetName, "repository", repoName, "group", groupName, "err", err)
	}
	return err
}

func artifactoryDeleteRepoPerms(orgName string) error {
	client, err := artifactoryClient()
	if err != nil {
		return err
	}
	permTargetName := getArtifactoryName(orgName)
	_, _, err = client.V1.Security.DeletePermissionTargets(context.Background(), permTargetName)
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
