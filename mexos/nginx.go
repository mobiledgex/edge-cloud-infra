package mexos

import (
	"fmt"
	"strings"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/parnurzeal/gorequest"
)

func AddNginxProxy(mf *Manifest, rootLBName, name, ipaddr string, ports []PortDetail, network string) error {
	log.DebugLog(log.DebugLevelMexos, "add nginx proxy", "name", name, "network", network, "ports", ports)

	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx", rootLBName, mf.Values.Agent.Port)
	pl, err := FormNginxProxyRequest(ports, ipaddr, name, network)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot form nginx proxy request")
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "nginx proxy add request post", "request", *pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx proxy add, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "ok, nginx proxy add request post")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "warning, error while adding nginx proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot add nginx proxy, resp %v", resp)
}

func FormNginxKCProxyRequest(name string, portnum int) (*string, error) {
	pl := fmt.Sprintf(`{"message":"add","name": "%s","port": "%d"}`, name, portnum)
	return &pl, nil
}

func FormNginxProxyRequest(ports []PortDetail, ipaddr string, name string, network string) (*string, error) {
	portstrs := []string{}
	for _, p := range ports {
		switch p.MexProto {
		case "LProtoHTTP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "internal": "%d", "origin":"%s:%d", "path":"/%s"}`,
					p.MexProto, p.PublicPort, p.InternalPort, ipaddr, p.InternalPort, p.PublicPath))
		case "LProtoTCP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "origin": "%s:%d"}`,
					p.MexProto, p.PublicPort, ipaddr, p.InternalPort))
		case "LProtoUDP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "origin": "%s:%d"}`,
					p.MexProto, p.PublicPort, ipaddr, p.InternalPort))
		default:
			log.DebugLog(log.DebugLevelMexos, "invalid mexproto", "port", p)
		}
	}
	portspec := ""
	for i, ps := range portstrs {
		if i == 0 {
			portspec += ps
		} else {
			portspec += "," + ps
		}

	}
	pl := fmt.Sprintf(`{ "message":"add", "name": "%s", "network": "%s", "ports": %s }`, name, network, "["+portspec+"]")
	if network != "" { //TODO: network is not handled right, and also incorrect in mexosagent handler

	}

	return &pl, nil
}

func DeleteNginxProxy(mf *Manifest, rootLBName, name string) error {
	log.DebugLog(log.DebugLevelMexos, "delete nginx proxy", "name", name)
	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx", rootLBName, mf.Values.Agent.Port)
	pl := fmt.Sprintf(`{"message":"delete","name":"%s"}`, name)
	log.DebugLog(log.DebugLevelMexos, "nginx proxy delete request post", "request", pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx proxy delete, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "deleted nginx proxy OK")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "error while deleting nginx proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot delete nginx proxy, resp %v", resp)
}

func AddNginxKubectlProxy(mf *Manifest, rootLBName, name string, portnum int) error {
	log.DebugLog(log.DebugLevelMexos, "add nginx kubectl proxy", "name", name)
	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx-kcp", rootLBName, mf.Values.Agent.Port)
	pl, err := FormNginxKCProxyRequest(name, portnum)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot form nginx kubectl proxy request")
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "nginx kubectl proxy add request post", "request", *pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx kubectl proxy add, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "ok, nginx kubectl proxy add request post")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "warning, error while adding nginx kubectl proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot add nginx kubectl proxy, resp %v", resp)
}

func DeleteNginxKCProxy(mf *Manifest, rootLBName, name string) error {
	log.DebugLog(log.DebugLevelMexos, "deleting nginx kubectl proxy", "name", name)
	out, err := sh.Command("docker", "kill", name+kcproxySuffix).Output()
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "warning, cannot delete container", "name", name+kcproxySuffix, "error", err, "out", out)
	}
	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx-kcp", rootLBName, mf.Values.Agent.Port)
	pl := fmt.Sprintf(`{"message":"delete","name":"%s"}`, name)
	log.DebugLog(log.DebugLevelMexos, "nginx kubectl proxy delete request post", "request", pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx kubectl proxy delete, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "deleted nginx kubectl proxy OK")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "error while deleting nginx kubectl proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot delete nginx kubectl proxy, resp %v", resp)
}
