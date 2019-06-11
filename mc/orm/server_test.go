package orm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	addr := "127.0.0.1:9999"
	uri := "http://" + addr + "/api/v1"

	config := ServerConfig{
		ServAddr:        addr,
		SqlAddr:         "127.0.0.1:5445",
		RunLocal:        true,
		InitLocal:       true,
		IgnoreEnv:       true,
		SkipVerifyEmail: true,
	}
	server, err := RunServer(&config)
	require.Nil(t, err, "run server")
	defer server.Stop()

	Jwks.Init("addr", "region", "mcorm", "roleID", "secretID")
	Jwks.Meta.CurrentVersion = 1
	Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}

	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	mcClient := &ormclient.Client{}

	// login as super user
	token, err := mcClient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	super, status, err := showCurrentUser(mcClient, uri, token)
	require.Nil(t, err, "show super")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, DefaultSuperuser, super.Name, "super user name")
	require.Equal(t, "", super.Passhash, "empty pass hash")
	require.Equal(t, "", super.Salt, "empty salt")
	require.Equal(t, 0, super.Iter, "empty iter")

	roleAssignments, status, err := mcClient.ShowRoleAssignment(uri, token)
	require.Nil(t, err, "show roles")
	require.Equal(t, http.StatusOK, status, "show role status")
	require.Equal(t, 1, len(roleAssignments), "num role assignments")
	require.Equal(t, RoleAdminManager, roleAssignments[0].Role)
	require.Equal(t, super.Name, roleAssignments[0].Username)

	// show users - only super user at this point
	users, status, err := mcClient.ShowUser(uri, token, &ormapi.Organization{})
	require.Equal(t, http.StatusOK, status, "show user status")
	require.Equal(t, 1, len(users))
	require.Equal(t, DefaultSuperuser, users[0].Name, "super user name")
	require.Equal(t, "", users[0].Passhash, "empty pass hash")
	require.Equal(t, "", users[0].Salt, "empty salt")
	require.Equal(t, 0, users[0].Iter, "empty iter")

	policies, status, err := showRolePerms(mcClient, uri, token)
	require.Nil(t, err, "show role perms err")
	require.Equal(t, http.StatusOK, status, "show role perms status")
	require.Equal(t, 99, len(policies), "number of role perms")
	roles, status, err := showRoles(mcClient, uri, token)
	require.Nil(t, err, "show roles err")
	require.Equal(t, http.StatusOK, status, "show roles status")
	require.Equal(t, 9, len(roles), "number of roles")

	// create new user1
	user1 := ormapi.User{
		Name:     "MisterX",
		Email:    "misterx@gmail.com",
		Passhash: "misterx-password",
	}
	status, err = mcClient.CreateUser(uri, &user1)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user1
	tokenMisterX, err := mcClient.DoLogin(uri, user1.Name, user1.Passhash)
	require.Nil(t, err, "login as mister X")
	// create an Organization
	org1 := ormapi.Organization{
		Type:    "developer",
		Name:    "DevX",
		Address: "123 X Way",
		Phone:   "123-123-1234",
	}
	status, err = mcClient.CreateOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// create new user2
	user2 := ormapi.User{
		Name:     "MisterY",
		Email:    "mistery@gmail.com",
		Passhash: "mistery-password",
	}
	status, err = mcClient.CreateUser(uri, &user2)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user2
	tokenMisterY, err := mcClient.DoLogin(uri, user2.Name, user2.Passhash)
	require.Nil(t, err, "login as mister Y")
	// create an Organization
	org2 := ormapi.Organization{
		Type:    "developer",
		Name:    "DevY",
		Address: "123 Y Way",
		Phone:   "123-321-1234",
	}
	status, err = mcClient.CreateOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// create new admin user
	admin := ormapi.User{
		Name:     "Admin",
		Email:    "Admin@gmail.com",
		Passhash: "admin-password",
	}
	status, err = mcClient.CreateUser(uri, &admin)
	require.Nil(t, err, "create admin user")
	require.Equal(t, http.StatusOK, status, "create admin user status")
	// add admin user as admin role
	roleArg := ormapi.Role{
		Username: admin.Name,
		Role:     "AdminManager",
	}
	status, err = mcClient.AddUserRole(uri, token, &roleArg)
	require.Nil(t, err, "add user role")
	require.Equal(t, http.StatusOK, status)
	// login as new admin
	tokenAdmin, err := mcClient.DoLogin(uri, admin.Name, admin.Passhash)
	require.Nil(t, err, "login as admin")

	// check org membership as mister x
	orgs, status, err := mcClient.ShowOrg(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org1.Name, orgs[0].Name)
	require.Equal(t, org1.Type, orgs[0].Type)
	// check org membership as mister y
	orgs, status, err = mcClient.ShowOrg(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org2.Name, orgs[0].Name)
	require.Equal(t, org2.Type, orgs[0].Type)
	// super user should be able to show all orgs
	orgs, status, err = mcClient.ShowOrg(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(orgs))
	orgs, status, err = mcClient.ShowOrg(uri, tokenAdmin)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(orgs))

	// check role assignments as mister x
	roleAssignments, status, err = mcClient.ShowRoleAssignment(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user1.Name, roleAssignments[0].Username)
	// check role assignments as mister y
	roleAssignments, status, err = mcClient.ShowRoleAssignment(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user2.Name, roleAssignments[0].Username)
	// super user should be able to see all role assignments
	roleAssignments, status, err = mcClient.ShowRoleAssignment(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 4, len(roleAssignments))
	roleAssignments, status, err = mcClient.ShowRoleAssignment(uri, tokenAdmin)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 4, len(roleAssignments))

	// show org users as mister x
	users, status, err = mcClient.ShowUser(uri, tokenMisterX, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user1.Name, users[0].Name)
	// show org users as mister y
	users, status, err = mcClient.ShowUser(uri, tokenMisterY, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user2.Name, users[0].Name)
	// super user can see all users with org ID = 0
	users, status, err = mcClient.ShowUser(uri, token, &ormapi.Organization{})
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 4, len(users))
	users, status, err = mcClient.ShowUser(uri, tokenAdmin, &ormapi.Organization{})
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 4, len(users))

	// check that x and y cannot see each other's org users
	users, status, err = mcClient.ShowUser(uri, tokenMisterX, &org2)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = mcClient.ShowUser(uri, tokenMisterY, &org1)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	foobar := &ormapi.Organization{
		Name: "foobar",
	}
	users, status, err = mcClient.ShowUser(uri, tokenMisterX, foobar)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// check that x and y cannot delete each other's orgs
	status, err = mcClient.DeleteOrg(uri, tokenMisterX, &org2)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	status, err = mcClient.DeleteOrg(uri, tokenMisterY, &org1)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// create more users
	user3, token3 := testCreateUser(t, mcClient, uri, "user3")
	user4, token4 := testCreateUser(t, mcClient, uri, "user4")
	// add users to org with different roles, make sure they can see users
	testAddUserRole(t, mcClient, uri, tokenMisterX, org1.Name, "DeveloperContributor", user3.Name, Success)
	testAddUserRole(t, mcClient, uri, tokenMisterX, org1.Name, "DeveloperViewer", user4.Name, Success)
	// check that they can see all users in org
	users, status, err = mcClient.ShowUser(uri, token3, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(users))
	users, status, err = mcClient.ShowUser(uri, token4, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(users))
	// make sure they can't see users from other org
	users, status, err = mcClient.ShowUser(uri, token3, &org2)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = mcClient.ShowUser(uri, token4, &org2)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// normal user cannot remove admin roles from others
	status, err = mcClient.RemoveUserRole(uri, tokenMisterX, &roleArg)
	require.NotNil(t, err, "remove user role")
	require.Equal(t, http.StatusForbidden, status)
	// admin user can remove role
	status, err = mcClient.RemoveUserRole(uri, tokenAdmin, &roleArg)
	require.Nil(t, err, "remove user role")
	require.Equal(t, http.StatusOK, status)

	// delete orgs
	status, err = mcClient.DeleteOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	// delete users
	status, err = mcClient.DeleteUser(uri, tokenMisterX, &user1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteUser(uri, tokenMisterY, &user2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteUser(uri, token3, user3)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteUser(uri, token4, user4)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteUser(uri, tokenAdmin, &admin)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// check orgs are gone
	orgs, status, err = mcClient.ShowOrg(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(orgs))
	// check users are gone
	users, status, err = mcClient.ShowUser(uri, token, &ormapi.Organization{})
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))

	testLockedUsers(t, uri, mcClient)
}

func showCurrentUser(mcClient *ormclient.Client, uri, token string) (*ormapi.User, int, error) {
	user := ormapi.User{}
	status, err := mcClient.PostJson(uri+"/auth/user/current", token, nil, &user)
	return &user, status, err
}

func showRolePerms(mcClient *ormclient.Client, uri, token string) ([]ormapi.RolePerm, int, error) {
	perms := []ormapi.RolePerm{}
	status, err := mcClient.PostJson(uri+"/auth/role/perms/show", token, nil, &perms)
	return perms, status, err
}

func showRoles(mcClient *ormclient.Client, uri, token string) ([]string, int, error) {
	roles := []string{}
	status, err := mcClient.PostJson(uri+"/auth/role/show", token, nil, &roles)
	return roles, status, err
}

func waitServerOnline(addr string) error {
	return fmt.Errorf("wait server online failed")
}

// for debug
func dumpTables() {
	users := []ormapi.User{}
	orgs := []ormapi.Organization{}
	db.Find(&users)
	db.Find(&orgs)
	for _, user := range users {
		fmt.Printf("%v\n", user)
	}
	for _, org := range orgs {
		fmt.Printf("%v\n", org)
	}
}

func testLockedUsers(t *testing.T, uri string, mcClient *ormclient.Client) {
	// login as super user
	superTok, err := mcClient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	// set config to be locked. This needs to be a map so that
	// marshalling the JSON doesn't put in null entries for other
	// fields, and preserves null entries for specified fields regardless
	// of omit empty.
	notifyEmail := "foo@gmail.com"
	configReq := make(map[string]interface{})
	configReq["locknewaccounts"] = true
	configReq["notifyemailaddress"] = notifyEmail
	status, err := mcClient.UpdateConfig(uri, superTok, configReq)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// create new account
	user1 := ormapi.User{
		Name:     "user1",
		Email:    "user1@gmail.com",
		Passhash: "user1-password",
	}
	status, err = mcClient.CreateUser(uri, &user1)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user1
	_, err = mcClient.DoLogin(uri, user1.Name, user1.Passhash)
	require.NotNil(t, err, "login")
	require.Contains(t, err.Error(), "Account is locked")

	// super user unlock account
	userReq := make(map[string]interface{})
	userReq["email"] = user1.Email
	userReq["locked"] = false
	status, err = mcClient.RestrictedUserUpdate(uri, superTok, userReq)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// user should be able to log in now
	tok1, err := mcClient.DoLogin(uri, user1.Name, user1.Passhash)
	require.Nil(t, err)

	// create another new user
	user2 := ormapi.User{
		Name:     "user2",
		Email:    "user2@gmail.com",
		Passhash: "user2-password",
	}
	status, err = mcClient.CreateUser(uri, &user2)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user2
	_, err = mcClient.DoLogin(uri, user2.Name, user2.Passhash)
	require.NotNil(t, err, "login")
	require.Contains(t, err.Error(), "Account is locked")

	// make sure users cannot unlock other users
	userReq = make(map[string]interface{})
	userReq["email"] = user2.Email
	userReq["locked"] = false
	status, err = mcClient.RestrictedUserUpdate(uri, tok1, userReq)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// make sure users cannot modify config
	configReq = make(map[string]interface{})
	configReq["locknewaccounts"] = false
	status, err = mcClient.UpdateConfig(uri, tok1, configReq)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	// super user unlock new accounts
	status, err = mcClient.UpdateConfig(uri, superTok, configReq)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// user2 still should not be able to log in
	_, err = mcClient.DoLogin(uri, user2.Name, user2.Passhash)
	require.NotNil(t, err, "login")
	require.Contains(t, err.Error(), "Account is locked")

	// delete user2, recreate, should be unlocked now
	status, err = mcClient.DeleteUser(uri, superTok, &user2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.CreateUser(uri, &user2)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user2
	_, err = mcClient.DoLogin(uri, user2.Name, user2.Passhash)
	require.Nil(t, err)

	// show config, make sure changes didn't affect notify email address
	config, status, err := mcClient.ShowConfig(uri, superTok)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, notifyEmail, config.NotifyEmailAddress)

	// make sure users can't see config
	_, status, err = mcClient.ShowConfig(uri, tok1)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// clean up
	status, err = mcClient.DeleteUser(uri, superTok, &user1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = mcClient.DeleteUser(uri, superTok, &user2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}
