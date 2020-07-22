package orm

import (
	fmt "fmt"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type AlertManagerContext struct {
	claims *UserClaims
}

// Create alert receiver api handler
// TODO - make the list generic with respect to what type of user creates an alert
func CreateAlertReceiver(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	log.SpanLog(ctx, log.DebugLevelInfo, "Create Alertmanager Receiver", "context", c, "clainms", claims)
	in := ormapi.AlertReceiver{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	in.User = claims.Username
	if in.Cloudlet.Organization == "" && in.AppInst.AppKey.Organization == "" {
		return setReply(c,
			fmt.Errorf("Either cloudlet, or app instance details have to be specified"), nil)
	}
	// set up labels
	labels := map[string]string{}
	// Check that user is allowed to access either of the orgs
	if in.Cloudlet.Organization != "" {
		if err := authorized(ctx, claims.Username, in.Cloudlet.Organization,
			ResourceAppAnalytics, ActionView); err != nil {
			return setReply(c, err, nil)
		}
		// add labes for the cloudlet
		labels[edgeproto.CloudletKeyTagOrganization] = in.Cloudlet.Organization
		if in.Cloudlet.Name != "" {
			labels[edgeproto.CloudletKeyTagName] = in.Cloudlet.Name
		}
	}
	if in.AppInst.AppKey.Organization != "" {
		if err := authorized(ctx, claims.Username, in.AppInst.AppKey.Organization,
			ResourceAppAnalytics, ActionView); err != nil {
			return setReply(c, err, nil)
		}
		// add labels for app instance
		labels[edgeproto.AppKeyTagOrganization] = in.AppInst.AppKey.Organization
		if in.AppInst.AppKey.Name != "" {
			labels[edgeproto.AppKeyTagName] = in.AppInst.AppKey.Name
		}
		if in.AppInst.AppKey.Version != "" {
			labels[edgeproto.AppKeyTagVersion] = in.AppInst.AppKey.Version
		}
		if in.AppInst.ClusterInstKey.CloudletKey.Name != "" {
			labels[edgeproto.CloudletKeyTagName] = in.AppInst.ClusterInstKey.CloudletKey.Name
		}
		if in.AppInst.ClusterInstKey.CloudletKey.Organization != "" {
			labels[edgeproto.CloudletKeyTagOrganization] = in.AppInst.ClusterInstKey.CloudletKey.Organization
		}
		if in.AppInst.ClusterInstKey.ClusterKey.Name != "" {
			labels[edgeproto.ClusterKeyTagName] = in.AppInst.ClusterInstKey.ClusterKey.Name
		}
		if in.AppInst.ClusterInstKey.Organization != "" {
			labels[edgeproto.ClusterInstKeyTagOrganization] = in.AppInst.ClusterInstKey.Organization
		}
	}

	switch in.Type {
	case alertmgr.AlertReceiverTypeEmail:
		user := ormapi.User{
			Name:  claims.Username,
			Email: claims.Email,
		}
		err = AlertManagerServer.CreateReceiver(ctx, &in, labels, &user)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create a receiver", "err", err)
			return setReply(c, fmt.Errorf("Unable to create a receiver - %s", err.Error()),
				nil)
		}
	case alertmgr.AlertReceiverTypeSlack:
		// TODO
	default:
		log.SpanLog(ctx, log.DebugLevelInfo, "type of a receiver is invalid")
		return setReply(c, fmt.Errorf("Receiver type invalid"), nil)
	}
	return nil
}

// Delete alert receiver api handler
// TODO
func DeleteAlertReceiver(c echo.Context) error {
	return nil
}

// Show alert receivers api handler
// TODO
func ShowAlertReceiver(c echo.Context) error {
	return nil
}
