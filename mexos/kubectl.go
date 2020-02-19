package mexos

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
	v1 "k8s.io/api/core/v1"
)

func CreateDockerRegistrySecret(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, vaultConfig *vault.Config, names *k8smgmt.KubeNames) error {
	var out string
	log.SpanLog(ctx, log.DebugLevelMexos, "creating docker registry secret in kubernetes cluster")
	auth, err := cloudcommon.GetRegistryAuth(ctx, app.ImagePath, vaultConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "warning, cannot get docker registry secret from vault - assume public registry", "err", err)
		return nil
	}
	if auth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("auth type for %s is not basic auth type", auth.Hostname)
	}
	// Note: docker-server must contain port if imagepath contains port,
	// otherwise imagepullsecrets won't work.
	// Also secret name includes port in case multiple docker registries
	// are running on different ports on the same host.
	secretName := auth.Hostname
	dockerServer := auth.Hostname
	if auth.Port != "" {
		secretName = auth.Hostname + "-" + auth.Port
		dockerServer = auth.Hostname + ":" + auth.Port
	}
	// Note that the registry secret name must be per-app, since a developer
	// may put multiple apps in the same ClusterInst and they may come
	// from different registries.
	cmd := fmt.Sprintf("kubectl create secret docker-registry %s "+
		"--docker-server=%s --docker-username=%s --docker-password=%s "+
		"--docker-email=mobiledgex@mobiledgex.com --kubeconfig=%s",
		secretName, dockerServer, auth.Username, auth.Password,
		k8smgmt.GetKconfName(clusterInst))
	log.SpanLog(ctx, log.DebugLevelMexos, "CreateDockerRegistrySecret", "cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "AlreadyExists") {
			return fmt.Errorf("can't add docker registry secret, %s, %v", out, err)
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "warning, docker registry secret already exists.")
		}
	}
	names.ImagePullSecret = secretName
	log.SpanLog(ctx, log.DebugLevelMexos, "ok, created registry secret", "out", out)
	return nil
}

// ConfigMap of cluster instance details such as cluster name, cloudlet name, and operator name
func CreateClusterConfigMap(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error {
	var out string

	log.SpanLog(ctx, log.DebugLevelMexos, "creating cluster config map in kubernetes cluster")

	cmd := fmt.Sprintf("kubectl create configmap mexcluster-info "+
		"--from-literal=ClusterName='%s' "+
		"--from-literal=CloudletName='%s' "+
		"--from-literal=OperatorName='%s' --kubeconfig=%s",
		clusterInst.Key.ClusterKey.Name, clusterInst.Key.CloudletKey.Name,
		clusterInst.Key.CloudletKey.OperatorKey.Name,
		k8smgmt.GetKconfName(clusterInst))

	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "AlreadyExists") {
			return fmt.Errorf("can't add cluster ConfigMap cmd %s, %s, %v", cmd, out, err)
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "warning, Cluster ConfigMap already exists.")
		}
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "ok, created mexcluster-info configmap")
	return nil
}

func GetSvcExternalIP(ctx context.Context, client ssh.Client, kubeNames *k8smgmt.KubeNames, name string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "get service external IP", "name", name)
	externalIP := ""
	//wait for Load Balancer to assign external IP address. It takes a variable amount of time.
	for i := 0; i < 100; i++ {
		cmd := fmt.Sprintf("%s kubectl get svc -o json", kubeNames.KconfEnv)
		out, err := client.Output(cmd)
		if err != nil {
			return "", fmt.Errorf("error getting svc %s, %s, %v", name, out, err)
		}
		svcs, err := GetServices(ctx, client, kubeNames)
		if err != nil {
			return "", err
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "getting externalIP, examine list of services", "name", name, "svcs", svcs)
		for _, svc := range svcs {
			log.SpanLog(ctx, log.DebugLevelMexos, "svc item", "item", svc, "name", name)
			if svc.ObjectMeta.Name != name {
				log.SpanLog(ctx, log.DebugLevelMexos, "service name mismatch", "name", name, "svc.ObjectMeta.Name", svc.ObjectMeta.Name)
				continue
			}
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				log.SpanLog(ctx, log.DebugLevelMexos, "found ingress ip", "ingress.IP", ingress.IP, "svc.ObjectMeta.Name", svc.ObjectMeta.Name)
				if ingress.IP != "" {
					externalIP = ingress.IP
					log.SpanLog(ctx, log.DebugLevelMexos, "got externaIP for app", "externalIP", externalIP)
					return externalIP, nil
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	if externalIP == "" {
		return "", fmt.Errorf("timed out trying to get externalIP")
	}
	return externalIP, nil
}

func GetServices(ctx context.Context, client ssh.Client, names *k8smgmt.KubeNames) ([]v1.Service, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "get services", "kconf", names.KconfName)
	svcs := svcItems{}
	if names.DeploymentType == cloudcommon.AppDeploymentTypeDocker {
		// just populate the service names
		for _, sn := range names.ServiceNames {
			item := v1.Service{}
			item.Name = sn
			svcs.Items = append(svcs.Items, item)
		}
		return svcs.Items, nil
	}
	cmd := fmt.Sprintf("%s kubectl get svc -o json", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("can not get list of services: %s, %s, %v", cmd, out, err)
	}
	err = json.Unmarshal([]byte(out), &svcs)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "cannot unmarshal svc json", "out", out, "err", err)
		return nil, fmt.Errorf("cannot unmarshal svc json, %s", err.Error())
	}
	return svcs.Items, nil
}

func BackupKubeconfig(ctx context.Context, client ssh.Client) {
	kc := DefaultKubeconfig()
	cmd := fmt.Sprintf("mv %s %s.save", kc, kc)
	out, err := client.Output(cmd)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "can't rename", "name", kc, "err", err, "out", out)
	}
}

func parseKCPort(ln string) int {
	if !strings.Contains(ln, "kubectl") {
		return 0
	}
	if !strings.Contains(ln, "--port") {
		return 0
	}
	var a, b, c, port string
	n, serr := fmt.Sscanf(ln, "%s %s %s %s", &a, &b, &c, &port)
	if serr != nil {
		return 0
	}
	if n != 4 {
		return 0
	}
	portnum, aerr := strconv.Atoi(port)
	if aerr != nil {
		return 0
	}
	return portnum
}

func parseKCPid(ln string, key string) int {
	ln = strings.TrimSpace(ln)
	if !strings.Contains(ln, "kubectl") {
		return 0
	}
	if !strings.HasSuffix(ln, key) {
		return 0
	}
	var pid string
	n, serr := fmt.Sscanf(ln, "%s", &pid)
	if serr != nil {
		return 0
	}
	if n != 1 {
		return 0
	}
	pidnum, aerr := strconv.Atoi(pid)
	if aerr != nil {
		return 0
	}
	return pidnum
}
