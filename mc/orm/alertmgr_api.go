package orm

import (
	fmt "fmt"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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
	// Check that user is allowed to access either of the orgs
	if in.Cloudlet.Organization != "" {
		if err := authorized(ctx, claims.Username, in.Cloudlet.Organization,
			ResourceAlert, ActionView); err != nil {
			return setReply(c, err, nil)
		}
	}
	if in.AppInst.AppKey.Organization != "" {
		if err := authorized(ctx, claims.Username, in.AppInst.AppKey.Organization,
			ResourceAlert, ActionView); err != nil {
			return setReply(c, err, nil)
		}
	}

	switch in.Type {
	case alertmgr.AlertReceiverTypeEmail:
		user := ormapi.User{
			Name:  claims.Username,
			Email: claims.Email,
		}
		err = AlertManagerServer.CreateReceiver(ctx, &in, &user)
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
func DeleteAlertReceiver(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	log.SpanLog(ctx, log.DebugLevelInfo, "Delete Alertmanager Receiver", "context", c, "clainms", claims)
	in := ormapi.AlertReceiver{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	in.User = claims.Username
	// Since we actually use claims.Username, don't need to in fact authorize as the receivers are unique
	err = AlertManagerServer.DeleteReceiver(ctx, &in)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to delete a receiver", "err", err)
		return setReply(c, fmt.Errorf("Unable to delete a receiver - %s", err.Error()),
			nil)
	}
	return nil
}

// Show alert receivers api handler
func ShowAlertReceiver(c echo.Context) error {
	log.DebugLog(log.DebugLevelApi, "Running Show Alerts API")
	alertRecs := []ormapi.AlertReceiver{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	log.SpanLog(ctx, log.DebugLevelApi, "Show Alertmanager Receivers", "context", c, "clainms", claims)

	receivers, err := AlertManagerServer.ShowReceivers(ctx, nil)
	if err != nil {
		return err
	}
	for ii := range receivers {
		if receivers[ii].Cloudlet.Organization != "" {
			if err := authorized(ctx, claims.Username, receivers[ii].Cloudlet.Organization,
				ResourceAlert, ActionView); err == nil {
				alertRecs = append(alertRecs, receivers[ii])
			}
		} else {
			if err := authorized(ctx, claims.Username, receivers[ii].AppInst.AppKey.Organization,
				ResourceAlert, ActionView); err == nil {
				alertRecs = append(alertRecs, receivers[ii])
			}
		}
	}
	return setReply(c, err, alertRecs)
}
