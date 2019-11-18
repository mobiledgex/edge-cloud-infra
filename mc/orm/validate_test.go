package orm

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/stretchr/testify/require"
)

func TestValidName(t *testing.T) {
	var err error

	gname := GitlabGroupSanitize("niantic, inc.")
	require.Equal(t, "niantic--inc", gname)

	err = ValidName(".orgname_123.dev")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("-orgname_123.dev")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123.dev.")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123dev-cache")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123dev.git")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123dev.atom")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123dev test")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("orgname_123dev,test")
	require.NotNil(t, err, "invalid org name")

	err = ValidName("username_123dev::test")
	require.NotNil(t, err, "invalid user name")

	err = ValidName("username_123dev&test")
	require.NotNil(t, err, "invalid user name")

	err = ValidName(cloudcommon.DeveloperMobiledgeX)
	require.NotNil(t, err, "invalid org name")
}
