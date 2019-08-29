package mexos

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

const HTPasswdFile = "nginx.htpasswd"

func PrivateSSHKey() string {
	return MEXDir() + "/id_rsa_mex"
}

func MEXDir() string {
	return os.Getenv("HOME") + "/.mobiledgex"
}

func DefaultKubeconfig() string {
	return os.Getenv("HOME") + "/.kube/config"
}

func CopyFile(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SeedDockerSecret(ctx context.Context, client pc.PlatformClient, inst *edgeproto.ClusterInst, app *edgeproto.App, vaultAddr string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "seed docker secret")

	auth, err := cloudcommon.GetRegistryAuth(app.ImagePath, vaultAddr)
	if err != nil {
		return err
	}
	if auth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("auth type for %s is not basic auth type", auth.Hostname)
	}

	// XXX: not sure writing password to file buys us anything if the
	// echo command is recorded in some history.
	cmd := fmt.Sprintf("echo %s > .docker-pass", auth.Password)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker password, %s, %v", out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "stored docker password")
	defer func() {
		cmd := fmt.Sprintf("rm .docker-pass")
		client.Output(cmd)
	}()

	cmd = fmt.Sprintf("cat .docker-pass | docker login -u %s --password-stdin %s ", auth.Username, auth.Hostname)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't docker login on rootlb to %s, %s, %v", auth.Hostname, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "docker login ok")
	return nil
}

func GetHTPassword(ctx context.Context, rootLBName string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "get htpasswd")
	client, err := GetSSHClient(ctx, rootLBName, GetCloudletExternalNetwork(), SSHUser)
	if err != nil {
		return fmt.Errorf("can't get ssh client for docker swarm, %v", err)
	}
	cmd := fmt.Sprintf("scp -o %s -o %s -i id_rsa_mex mobiledgex@%s:files-repo/mobiledgex/%s .", sshOpts[0], sshOpts[1], GetCloudletRegistryFileServer(), HTPasswdFile)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't get htpasswd file, %v, %s", err, out)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "downloaded htpasswd")
	return nil
}
