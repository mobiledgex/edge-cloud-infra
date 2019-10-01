package orm

import (
	"context"
	"fmt"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestCheckImagePath(t *testing.T) {
	log.InitTracer("")
	ctx := log.StartTestSpan(context.Background())

	// non-mobiledgex paths always succeed
	testImagePath(t, ctx, "org1", "http://foobar.com/blah/blah", true)
	testImagePath(t, ctx, "org1", "http://foobar.com/artifactory/repo-blah/blah", true)
	// mobiledgex paths that should succeed - gitlab
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org1/app", true)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org1/app:1.0", true)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org1/app:latest", true)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org1/extra/app:latest", true)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org1/", true)
	// mobiledgex paths that should succeed - artifactory
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/artifactory/repo-org1/cirros-0.4.0-arm-disk.img#md5:7e9cfcb763e83573a4b9d9315f56cc5f", true)

	// should fail
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org2/app", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org2/app:1.0", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/org2/app:latest", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/artifactory/repo-org2/cirros-0.4.0-arm-disk.img#md5:7e9cfcb763e83573a4b9d9315f56cc5f", false)
	testImagePath(t, ctx, "org1", "http://foobar.mobiledgex.net/artifactory/org1/cirros-0.4.0-arm-disk.img#md5:7e9cfcb763e83573a4b9d9315f56cc5f", false)
}

func testImagePath(t *testing.T, ctx context.Context, org, imagepath string, ok bool) {
	app := edgeproto.App{}
	app.Key.DeveloperKey.Name = org
	app.ImagePath = imagepath
	err := checkImagePath(ctx, &app)
	if ok {
		require.Nil(t, err)
	} else {
		require.NotNil(t, err)
		fmt.Printf("%v\n", err)
	}
}
