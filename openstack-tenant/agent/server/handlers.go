package server

import (
	"fmt"

	log "gitlab.com/bobbae/logrus"

	"github.com/codeskyblue/go-sh"
	"golang.org/x/net/context"

	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/k8sopenstack"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/api"

	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//Server for GRPC
type Server struct{}

var proxyMap map[string]string

func init() {
	proxyMap = make(map[string]string)
}

//Provision a kubernetes cluster on openstack for the given Tenant
func (srv *Server) Provision(ctx context.Context, req *api.ProvisionRequest) (res *api.ProvisionResponse, err error) {
	if len(req.Provisions) < 1 {
		return nil, fmt.Errorf("missing provisions")
	}
	if err := k8sopenstack.CreateKubernetesCluster(req.Provisions[0]); err != nil { ///XXX only one for now
		return nil, fmt.Errorf("can't create kubernetes cluster on openstack, %v", err)
	}

	res = &api.ProvisionResponse{
		Message: req.Message,
		Status:  "OK",
	}

	return res, nil
}

//Destroy a kubernetes cluster on openstack for the given Tenant
func (srv *Server) Destroy(ctx context.Context, req *api.ProvisionRequest) (res *api.ProvisionResponse, err error) {
	if len(req.Provisions) < 1 {
		return nil, fmt.Errorf("missing provisions")
	}
	if err := k8sopenstack.DeleteKubernetesCluster(req.Provisions[0]); err != nil { ///XXX only one for now
		return nil, fmt.Errorf("can't create kubernetes cluster on openstack, %v", err)
	}

	res = &api.ProvisionResponse{
		Message: req.Message,
		Status:  "ok",
	}

	return res, nil
}

//Proxy traffic betwwen kubernetes pods, via openstack subnet/network, iptables, and external network
func (srv *Server) Proxy(ctx context.Context, req *api.ProxyRequest) (res *api.ProxyResponse, err error) {
	if len(req.Proxies) < 1 {
		return nil, fmt.Errorf("missing proxy definitions")
	}
	if req.Message == "add" {
		//TODO layer 4 proxy
		for _, p := range req.Proxies {
			err := addOrigin(p.Path, p.Origin)
			if err != nil {
				res := &api.ProxyResponse{
					Message: fmt.Sprintf("Error, cannot add proxy %s %s, %v", p.Path, p.Origin, err),
				}
				return res, err
			}
		}
	} else {
		//TODO list
		res := &api.ProxyResponse{
			Message: fmt.Sprintf("Error, invalid request %s", req.Message),
		}
		return res, err
	}

	res = &api.ProxyResponse{
		Message: "OK",
	}

	return res, nil
}

func addOrigin(path, origin string) error {
	log.Debugf("addOrigin path %s origin %s", path, origin)

	originURL, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("Cannot parse origin %s,%v", origin, err)
	}

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originURL.Host)
		req.URL.Scheme = originURL.Scheme
		req.URL.Host = originURL.Host
		if strings.HasPrefix(req.URL.Path, path) {
			req.URL.Path = strings.Replace(req.URL.Path, path, "/", 1)
		} else {
			log.Warningf("invalid URL path %s missing %s", req.URL.Path, path)
		}
		log.Debugf("director req %v", req)
	}

	proxy := &httputil.ReverseProxy{Director: director}
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("path %s, request %v", path, r)
		proxy.ServeHTTP(w, r)
	})

	proxyMap[path] = origin
	return nil
}

//Status returns status information of the kubernetes cluster on openstack for a given tenant
func (srv *Server) Status(ctx context.Context, req *api.StatusRequest) (res *api.StatusResponse, err error) {
	res = &api.StatusResponse{
		Message: "OK",
	}

	return res, nil
}

func (srv *Server) Route(ctx context.Context, req *api.RouteRequest) (res *api.RouteResponse, err error) {
	if len(req.Routes) < 1 {
		return nil, fmt.Errorf("missing Route definitions")
	}
	if req.Message == "add" {
		for _, r := range req.Routes {
			err := addRoute(r.Subnet, r.Gateway, r.Dev)
			if err != nil {
				res := &api.RouteResponse{
					Message: fmt.Sprintf("Error, cannot add Route %s %s %s, %v", r.Subnet, r.Gateway, r.Dev, err),
				}
				return res, err
			}
		}
	} else if req.Message == "list" {
		rl, err := listRoutes()
		if err != nil {
			res := &api.RouteResponse{
				Message: fmt.Sprintf("Error, cannot list route, %v", err),
			}
			return res, err
		}

		res := &api.RouteResponse{
			Message: rl,
		}
		return res, nil
	} else {
		//TODO list
		res := &api.RouteResponse{
			Message: fmt.Sprintf("Error, invalid request %s", req.Message),
		}
		return res, err
	}

	res = &api.RouteResponse{
		Message: "OK",
	}

	return res, nil
}

func addRoute(subnet, gateway, dev string) error {
	log.Debugf("addRoute subnet %s gateway %s dev %s", subnet, gateway, dev)

	out, err := sh.Command("ip", "route", "add", subnet, "via", gateway, "dev", dev).CombinedOutput()
	if err != nil {
		return fmt.Errorf("can't add route %s %s %s, %s, %v", subnet, gateway, dev, out, err)
	}
	return nil
}

func listRoutes() (string, error) {
	log.Debugf("list routes")

	out, err := sh.Command("ip", "route", "show").Output()
	if err != nil {
		return "", fmt.Errorf("can't list route, %v", err)
	}
	return string(out), err
}
