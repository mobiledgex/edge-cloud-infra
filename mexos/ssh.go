package mexos

import (
	"fmt"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/nanobox-io/golang-ssh"
)

var sshOpts = []string{"StrictHostKeyChecking=no", "UserKnownHostsFile=/dev/null", "LogLevel=ERROR"}

//CopySSHCredential copies over the ssh credential for mex to LB
func CopySSHCredential(mf *Manifest, serverName, networkName, userName string) error {
	log.DebugLog(log.DebugLevelMexos, "copying ssh credentials", "server", serverName, "network", networkName, "user", userName)
	addr, err := GetServerIPAddr(mf, networkName, serverName)
	if err != nil {
		return err
	}
	kf := PrivateSSHKey()
	out, err := sh.Command("scp", "-o", sshOpts[0], "-o", sshOpts[1], "-i", kf, kf, "root@"+addr+":").Output()
	if err != nil {
		return fmt.Errorf("can't copy %s to %s, %s, %v", kf, addr, out, err)
	}
	return nil
}

//GetSSHClient returns ssh client handle for the server
func GetSSHClient(mf *Manifest, serverName, networkName, userName string) (ssh.Client, error) {
	auth := ssh.Auth{Keys: []string{PrivateSSHKey()}}
	addr, err := GetServerIPAddr(mf, networkName, serverName)
	if err != nil {
		return nil, err
	}
	client, err := ssh.NewNativeClient(userName, addr, "SSH-2.0-mobiledgex-ssh-client-1.0", 22, &auth, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get ssh client, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "got ssh client", "addr", addr, "key", auth)
	return client, nil
}
