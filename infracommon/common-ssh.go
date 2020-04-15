package infracommon

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/tmc/scp"
)

var DefaultConnectTimeout time.Duration = 30 * time.Second
var ClientVersion = "SSH-2.0-mobiledgex-ssh-client-1.0"

var SSHOpts = []string{"StrictHostKeyChecking=no", "UserKnownHostsFile=/dev/null", "LogLevel=ERROR"}
var SSHUser = "ubuntu"
var SSHPrivateKeyName = "id_rsa_mex"

type SSHOptions struct {
	Timeout time.Duration
	User    string
}

type SSHClientOp func(sshp *SSHOptions) error

func WithUser(user string) SSHClientOp {
	return func(op *SSHOptions) error {
		op.User = user
		return nil
	}
}

func WithTimeout(timeout time.Duration) SSHClientOp {
	return func(op *SSHOptions) error {
		op.Timeout = timeout
		return nil
	}
}

func (o *SSHOptions) Apply(ops []SSHClientOp) {
	for _, op := range ops {
		op(o)
	}
}

//CopySSHCredential copies over the ssh credential for mex to LB
func (c *CommonPlatform) CopySSHCredential(ctx context.Context, serverName, networkName, userName string) error {
	//TODO multiple keys to be copied and added to authorized_keys if needed
	log.SpanLog(ctx, log.DebugLevelInfra, "copying ssh credentials", "server", serverName, "network", networkName, "user", userName)
	ip, err := c.infraProvider.GetIPFromServerName(ctx, networkName, serverName)
	if err != nil {
		return err
	}
	kf := PrivateSSHKey()
	out, err := sh.Command("scp", "-o", SSHOpts[0], "-o", SSHOpts[1], "-i", kf, kf, userName+"@"+ip.ExternalAddr+":").Output()
	if err != nil {
		return fmt.Errorf("can't copy %s to %s, %s, %v", kf, ip.ExternalAddr, out, err)
	}
	return nil
}

//GetSSHClientFromIPAddr returns ssh client handle for the given IP.
func (c *CommonPlatform) GetSSHClientFromIPAddr(ctx context.Context, ipaddr string, ops ...SSHClientOp) (ssh.Client, error) {
	opts := SSHOptions{Timeout: DefaultConnectTimeout, User: SSHUser}
	opts.Apply(ops)
	var client ssh.Client
	var err error
	auth := ssh.Auth{Keys: []string{PrivateSSHKey()}}
	gwhost, gwport := c.GetCloudletCRMGatewayIPAndPort()
	if gwhost != "" {
		// start the client to GW and add the addr as next hop
		client, err = ssh.NewNativeClient(opts.User, ClientVersion, gwhost, gwport, &auth, opts.Timeout, nil)
		if err != nil {
			return nil, err
		}
		client, err = client.AddHop(ipaddr, 22)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		client, err = ssh.NewNativeClient(SSHUser, ClientVersion, ipaddr, 22, &auth, opts.Timeout, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot get ssh client for addr %s, %v", ipaddr, err)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Created SSH Client", "ipaddr", ipaddr, "gwhost", gwhost, "timeout", opts.Timeout)
	return client, nil
}

func (c *CommonPlatform) GetSSHClientForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := c.SharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(c.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return c.GetSSHClientForServer(ctx, rootLBName, c.GetCloudletExternalNetwork())
}

//GetSSHClient returns ssh client handle for the server
func (c *CommonPlatform) GetSSHClientForServer(ctx context.Context, serverName, networkName string, ops ...SSHClientOp) (ssh.Client, error) {
	// if this is a rootLB we may have the IP cached already
	var externalAddr string
	rootLB, err := c.GetRootLB(ctx, serverName)
	if err == nil && rootLB != nil {
		if rootLB.IP != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "using existing rootLB IP", "IP", rootLB.IP)
			externalAddr = rootLB.IP.ExternalAddr
		}
	}
	if externalAddr == "" {
		serverIp, err := c.infraProvider.GetIPFromServerName(ctx, networkName, serverName)
		if err != nil {
			return nil, err
		}
		externalAddr = serverIp.ExternalAddr
	}
	return c.GetSSHClientFromIPAddr(ctx, externalAddr, ops...)
}

func (c *CommonPlatform) SetupSSHUser(ctx context.Context, rootLB *MEXRootLB, user string) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "setting up ssh user", "user", user)
	client, err := c.GetSSHClientForServer(ctx, rootLB.Name, c.GetCloudletExternalNetwork(), WithUser(user))
	if err != nil {
		return nil, err
	}
	// XXX cloud-init creates non root user but it does not populate all the needed files.
	//  packer will create images with correct things for root .ssh. It cannot provision
	//  them for the `ubuntu` user. It may not yet exist until cloud-init runs. So we fix it here.
	for _, cmd := range []string{
		fmt.Sprintf("sudo cp /root/.ssh/config /home/%s/.ssh/", user),
		fmt.Sprintf("sudo chown %s:%s /home/%s/.ssh/config", user, user, user),
		fmt.Sprintf("sudo chmod 600 /home/%s/.ssh/config", user),
		fmt.Sprintf("sudo cp /root/%s /home/%s/", SSHPrivateKeyName, user),
		fmt.Sprintf("sudo chown %s:%s   /home/%s/%s", user, user, user, SSHPrivateKeyName),
		fmt.Sprintf("sudo chmod 600   /home/%s/%s", user, SSHPrivateKeyName),
	} {
		out, err := client.Output(cmd)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "error setting up ssh user",
				"user", user, "error", err, "out", out)
			return nil, err
		}
	}
	return client, nil
}

func SCPFilePath(sshClient ssh.Client, srcPath, dstPath string) error {
	client, ok := sshClient.(*ssh.NativeClient)
	if !ok {
		return fmt.Errorf("unable to cast client to native client")
	}
	session, sessionInfo, err := client.Session(client.DefaultClientConfig.Timeout)
	if err != nil {
		return err
	}
	defer sessionInfo.CloseAll()
	err = scp.CopyPath(srcPath, dstPath, session)
	return err
}
