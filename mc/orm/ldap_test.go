package orm

import (
	"testing"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
	"gopkg.in/ldap.v3"
)

func TestLDAPServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	addr := "127.0.0.1:9999"
	uri := "http://" + addr + "/api/v1"

	config := ServerConfig{
		ServAddr:  addr,
		SqlAddr:   "127.0.0.1:5445",
		RunLocal:  true,
		InitLocal: true,
		IgnoreEnv: true,
		LDAPAddr:  "127.0.0.1:9389",
	}
	server, err := RunServer(&config)
	require.Nil(t, err, "run server")
	defer server.Stop()

	Jwks.Init("addr", "mcorm", "roleID", "secretID")
	Jwks.Meta.CurrentVersion = 1
	Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}

	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	mcClient := &ormclient.Client{}

	// login as super user
	_, err = mcClient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	// create new users & orgs
	_, token1 := testCreateUser(t, mcClient, uri, "orgman1")
	org1 := testCreateOrg(t, mcClient, uri, token1, "developer", "bigorg1")
	worker1, _ := testCreateUser(t, mcClient, uri, "worker1")
	testAddUserRole(t, mcClient, uri, token1, org1.Name, "DeveloperContributor", worker1.Name, Success)

	_, token2 := testCreateUser(t, mcClient, uri, "orgman2")
	org2 := testCreateOrg(t, mcClient, uri, token2, "developer", "bigorg2")
	worker2, _ := testCreateUser(t, mcClient, uri, "worker2")
	testAddUserRole(t, mcClient, uri, token2, org2.Name, "DeveloperContributor", worker2.Name, Success)

	l, err := ldap.Dial("tcp", config.LDAPAddr)
	require.Nil(t, err, "connected to ldap server")
	defer l.Close()
	//l.debug = true

	// Expect Count: 7 (1 admin entry + 4 users + 2 orgs)
	ldapSearchCheck(t, l, "cn=worker1,ou=users", "worker1-password", "", "(objectClass=*)", 7)

	// Expect Count: 5 (1 admin entry + 4 users)
	ldapSearchCheck(t, l, "cn=worker1,ou=users", "worker1-password", "ou=users", "(objectClass=*)", 5)

	// Expect Count: 1 (1 user)
	ldapSearchCheck(t, l, "cn=worker1,ou=users", "worker1-password", "cn=orgman2,ou=users", "(objectClass=*)", 1)

	// Expect Count: 2 (2 orgs)
	ldapSearchCheck(t, l, "cn=orgman1,ou=users", "orgman1-password", "ou=orgs", "(objectClass=*)", 2)

	// Expect Count: 1 (1 org)
	ldapSearchCheck(t, l, "cn=orgman1,ou=users", "orgman1-password", "cn=bigorg1,ou=orgs", "(objectClass=*)", 1)

	// Expect Count: 1 (1 user)
	ldapSearchCheck(t, l, "cn=gitlab,ou=users", "gitlab", "", "(sAMAccountName=worker2)", 1)

	// Expect Count: 1 (2 orgs)
	ldapSearchCheck(t, l, "cn=gitlab,ou=users", "gitlab", "", "(objectClass=groupOfUniqueNames)", 2)

	// Expect Count: 1 (1 user)
	ldapSearchCheck(t, l, "cn=gitlab,ou=users", "gitlab", "ou=users", "(email=orgman1@gmail.com)", 1)

	// Expect Count: 1 (1 org)
	ldapSearchCheck(t, l, "cn=orgman1,ou=users", "orgman1-password", "ou=orgs", "(&(objectClass=groupOfUniqueNames)(|(uniqueMember=cn=worker2,ou=users)(uniqueMember=worker1)))", 1)
}

func ldapSearchCheck(t *testing.T, l *ldap.Conn, bindDN, bindPassword, baseDN, filter string, numEntries int) {

	err := l.Bind(bindDN, bindPassword)
	require.Nil(t, err, "ldap bind")

	searchRequest := &ldap.SearchRequest{
		BaseDN: baseDN,
		Filter: filter,
	}
	sr, err := l.Search(searchRequest)
	require.Nil(t, err, "ldap search")
	require.Equal(t, len(sr.Entries), numEntries, "match num of entries from search result")
}
