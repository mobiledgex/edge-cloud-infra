package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
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
}

type SSHClientOp func(sshp *SSHOptions) error

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

//GetSSHClient returns ssh client handle for the server
func (s *Platform) GetSSHClient(ctx context.Context, serverName, networkName, userName string, ops ...SSHClientOp) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "GetSSHClient", "serverName", serverName)
	opts := SSHOptions{Timeout: DefaultConnectTimeout}
	opts.Apply(ops)

	addr, err := s.GetServerIPAddr(ctx, networkName, serverName)
	if err != nil {
		return nil, err
	}

	var client ssh.Client
	var auth ssh.Auth
	if s.authKey != nil {
		auth.RawKeys = [][]byte{
			[]byte(s.authKey.PrivateKey),
		}
	} else {
		auth.Keys = []string{mexos.PrivateSSHKey()}
	}
	gwhost, gwport := s.GetCloudletCRMGatewayIPAndPort()
	if gwhost != "" {
		// start the client to GW and add the addr as next hop
		client, err = ssh.NewNativeClient(userName, ClientVersion, gwhost, gwport, &auth, opts.Timeout, nil)
		if err != nil {
			return nil, err
		}
		client, err = client.AddHop(addr.ExternalAddr, 22)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = ssh.NewNativeClient(userName, ClientVersion, addr.ExternalAddr, 22, &auth, opts.Timeout, nil)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Created SSH Client", "addr", addr, "gwhost", gwhost, "timeout", opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot get ssh client for server %s on network %s, %v", serverName, networkName, err)
	}
	//log.SpanLog(ctx,log.DebugLevelMexos, "got ssh client", "addr", addr, "key", auth)
	return client, nil
}

func (s *Platform) SetupSSHUser(ctx context.Context, rootLB *MEXRootLB, user string) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "setting up ssh user", "user", user)
	client, err := s.GetSSHClient(ctx, rootLB.Name, s.GetCloudletExternalNetwork(), user)
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
			log.SpanLog(ctx, log.DebugLevelMexos, "error setting up ssh user",
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
