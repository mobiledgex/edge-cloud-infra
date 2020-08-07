package gcp

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// SetProject sets the project in gcloud config
func (g *GCPPlatform) SetProject(project string) error {
	out, err := sh.Command("gcloud", "config", "set", "project", project).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

// SetZone sets the zone in gcloud config
func (g *GCPPlatform) SetZone(ctx context.Context, zone string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetZone", "zone", zone)
	out, err := sh.Command("gcloud", "config", "set", "compute/zone", zone).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

// CreateClusterPrerequisites currently does nothing
func (a *GCPPlatform) CreateClusterPrerequisites(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	return nil
}

// CreateGKECluster creates a kubernetes cluster on gcloud
func (g *GCPPlatform) RunClusterCreateCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	clusterName := g.NameSanitize(k8smgmt.GetClusterName(clusterInst))
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterCreateCommand", "clusterInst", clusterInst, "clusterName", clusterName)
	out, err := sh.Command("gcloud", "container", "clusters", "create", clusterName).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in cluster create", "out", out, "err", err)
		return fmt.Errorf("Error in cluster create: %s %v", out, err)
	}
	return nil
}

// RunClusterDeleteCommand removes kubernetes cluster on gcloud
func (g *GCPPlatform) RunClusterDeleteCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterDeleteCommand", "clusterInst", clusterInst)
	clusterName := g.NameSanitize(k8smgmt.GetClusterName(clusterInst))
	out, err := sh.Command("gcloud", "container", "clusters", "delete", "--quiet", clusterName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

// GetCredentials retrieves kubeconfig credentials from gcloud.
func (g *GCPPlatform) GetCredentials(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	clusterName := g.NameSanitize(k8smgmt.GetClusterName(clusterInst))
	out, err := sh.Command("gcloud", "container", "clusters", "get-credentials", clusterName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}
