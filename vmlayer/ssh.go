package vmlayer

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/tmc/scp"
)

var CloudletSSHKeyRefreshInterval = 24 * time.Hour

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

func (v *VMPlatform) SetCloudletSignedSSHKey(ctx context.Context, vaultConfig *vault.Config) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Sign cloudlet public key from Vault")

	data := map[string]interface{}{
		"public_key": v.VMProperties.sshKey.PublicKey,
	}
	signedKey, err := infracommon.GetSignedKeyFromVault(vaultConfig, data)
	if err != nil {
		return err
	}

	v.VMProperties.sshKey.Mux.Lock()
	defer v.VMProperties.sshKey.Mux.Unlock()
	v.VMProperties.sshKey.SignedPublicKey = signedKey

	return nil
}

func (v *VMPlatform) RefreshCloudletSSHKeys(vaultConfig *vault.Config) {
	interval := CloudletSSHKeyRefreshInterval
	for {
		select {
		case <-time.After(interval):
			span := log.StartSpan(log.DebugLevelInfra, "refresh Cloudlet SSH Key")
			ctx := log.ContextWithSpan(context.Background(), span)
			err := v.SetCloudletSignedSSHKey(ctx, vaultConfig)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "refresh cloudlet ssh key failure", "err", err)
				// retry again soon
				interval = time.Hour
			} else {
				interval = CloudletSSHKeyRefreshInterval
			}
			span.Finish()
		}
	}
}

func (v *VMPlatform) InitCloudletSSHKeys(ctx context.Context, vaultConfig *vault.Config) error {
	// Generate Cloudlet SSH Keys
	cloudletPubKey, cloudletPrivKey, err := ssh.GenKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate cloudlet SSH key pair: %v", err)
	}
	v.VMProperties.sshKey.PublicKey = cloudletPubKey
	v.VMProperties.sshKey.PrivateKey = cloudletPrivKey
	err = v.SetCloudletSignedSSHKey(ctx, vaultConfig)
	if err != nil {
		return err
	}
	return nil
}

//GetSSHClientFromIPAddr returns ssh client handle for the given IP.
func (vp *VMProperties) GetSSHClientFromIPAddr(ctx context.Context, ipaddr string, ops ...SSHClientOp) (ssh.Client, error) {
	opts := SSHOptions{Timeout: infracommon.DefaultConnectTimeout, User: infracommon.SSHUser}
	opts.Apply(ops)
	var client ssh.Client
	var err error

	if vp.sshKey.PrivateKey == "" {
		return nil, fmt.Errorf("missing cloudlet private key")
	}
	if vp.sshKey.SignedPublicKey == "" {
		return nil, fmt.Errorf("missing cloudlet signed public Key")
	}

	vp.sshKey.Mux.Lock()
	auth := ssh.Auth{
		KeyPairs: []ssh.KeyPair{
			ssh.KeyPair{
				PublicRawKey:  []byte(vp.sshKey.SignedPublicKey),
				PrivateRawKey: []byte(vp.sshKey.PrivateKey),
			},
		},
	}
	vp.sshKey.Mux.Unlock()

	gwhost, gwport := vp.GetCloudletCRMGatewayIPAndPort()
	if gwhost != "" {
		// start the client to GW and add the addr as next hop
		client, err = ssh.NewNativeClient(opts.User, infracommon.ClientVersion, gwhost, gwport, &auth, opts.Timeout, nil)
		if err != nil {
			return nil, err
		}
		client, err = client.AddHop(ipaddr, 22)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		client, err = ssh.NewNativeClient(infracommon.SSHUser, infracommon.ClientVersion, ipaddr, 22, &auth, opts.Timeout, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot get ssh client for addr %s, %v", ipaddr, err)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Created SSH Client", "ipaddr", ipaddr, "gwhost", gwhost, "timeout", opts.Timeout)
	return client, nil
}

func (v *VMPlatform) GetSSHClientForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := v.VMProperties.SharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(v.VMProperties.CommonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey, v.VMProperties.CommonPf.PlatformConfig.AppDNSRoot)
	}
	return v.GetSSHClientForServer(ctx, rootLBName, v.VMProperties.GetCloudletExternalNetwork())
}

//GetSSHClient returns ssh client handle for the server
func (v *VMPlatform) GetSSHClientForServer(ctx context.Context, serverName, networkName string, ops ...SSHClientOp) (ssh.Client, error) {
	// if this is a rootLB we may have the IP cached already
	var externalAddr string
	rootLB, err := GetRootLB(ctx, serverName)
	if err == nil && rootLB != nil {
		if rootLB.IP != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "using existing rootLB IP", "IP", rootLB.IP)
			externalAddr = rootLB.IP.ExternalAddr
		}
	}
	if externalAddr == "" {
		serverIp, err := v.GetIPFromServerName(ctx, networkName, "", serverName)
		if err != nil {
			return nil, err
		}
		externalAddr = serverIp.ExternalAddr
	}
	return v.VMProperties.GetSSHClientFromIPAddr(ctx, externalAddr, ops...)
}

func (v *VMPlatform) SetupSSHUser(ctx context.Context, rootLB *MEXRootLB, user string) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "setting up ssh user", "user", user)
	client, err := v.GetSSHClientForServer(ctx, rootLB.Name, v.VMProperties.GetCloudletExternalNetwork(), WithUser(user))
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
