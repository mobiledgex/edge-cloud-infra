package orm

import (
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func generateCloudletList(cloudletList []string) string {
	if len(cloudletList) == 0 {
		return ""
	}
	// format needs to be: cloudlet='cloudlet1' OR cloudlet='cloudlet2' ... OR cloudlet='cloudlet3'
	new := strings.Join(cloudletList, "' OR cloudlet='")
	new = "cloudlet='" + new + "'"
	return new
}

func SortCloudletPoolUsage(obj *ormapi.RegionCloudletPoolUsage, clusterUsage, appUsage *ormapi.AllUsage, cloudletList []string) (*ormapi.CloudletPoolUsage, error) {
	usage := ormapi.CloudletPoolUsage{
		Region:       obj.Region,
		CloudletPool: obj.CloudletPool.Name,
		Organization: obj.CloudletPool.Organization,
		StartTime:    obj.StartTime,
		EndTime:      obj.EndTime,
		Cloudlets:    []ormapi.CloudletUsage{},
	}
	cloudletMap := make(map[string]*ormapi.CloudletUsage)
	i := 0
	for _, cloudletName := range cloudletList {
		newCloudletUsage := ormapi.CloudletUsage{
			CloudletName: cloudletName,
			ClusterUsage: []ormapi.UsageRecord{},
			VmAppUsage:   []ormapi.UsageRecord{},
		}
		usage.Cloudlets = append(usage.Cloudlets, newCloudletUsage)
		cloudletMap[cloudletName] = &usage.Cloudlets[i]
		i = i + 1
	}
	for _, usageRecord := range clusterUsage.Data {
		record, ok := cloudletMap[usageRecord.Cloudlet]
		if !ok {
			return nil, fmt.Errorf("error sorting usage records")
		}
		record.ClusterUsage = append(record.ClusterUsage, usageRecord)
	}
	for _, usageRecord := range appUsage.Data {
		record, ok := cloudletMap[usageRecord.Cloudlet]
		if !ok {
			return nil, fmt.Errorf("error sorting usage records")
		}
		record.VmAppUsage = append(record.VmAppUsage, usageRecord)
	}
	return &usage, nil
}

func cloudletPoolEventsQuery(obj *ormapi.RegionCloudletPoolUsage, cloudletList []string, queryType string) string {
	arg := influxQueryArgs{
		OrgField:     "cloudletorg",
		ApiCallerOrg: obj.CloudletPool.Organization,
		CloudletList: generateCloudletList(cloudletList),
	}
	if queryType == CLUSTER {
		arg.Measurement = EVENT_CLUSTERINST
		arg.Selector = strings.Join(append(ClusterFields, clusterUsageEventFields...), ",")
	} else if queryType == APPINST {
		arg.Measurement = EVENT_APPINST
		arg.Selector = strings.Join(append(AppFields, appUsageEventFields...), ",")
		arg.DeploymentType = cloudcommon.DeploymentTypeKubernetes //TODO: change this to VM
	} else {
		return ""
	}
	queryStart := prevCheckpoint(obj.StartTime)
	return fillTimeAndGetCmd(&arg, usageInfluxDBTemplate, &queryStart, &obj.EndTime)
}

func cloudletPoolCheckpointsQuery(obj *ormapi.RegionCloudletPoolUsage, cloudletList []string, queryType string) string {
	arg := influxQueryArgs{
		OrgField:     "cloudletorg",
		ApiCallerOrg: obj.CloudletPool.Organization,
		CloudletList: generateCloudletList(cloudletList),
	}
	if queryType == CLUSTER {
		arg.Measurement = EVENT_CLUSTERINST
		arg.Selector = strings.Join(append(ClusterFields, clusterUsageEventFields...), ",")
	} else if queryType == APPINST {
		arg.Measurement = EVENT_APPINST
		arg.Selector = strings.Join(AppCheckpointFields, ",")
		arg.DeploymentType = cloudcommon.DeploymentTypeKubernetes //TODO: change this to VM
	} else {
		return ""
	}
	// set endtime to start and back up starttime by a checkpoint interval to hit the most recent
	// checkpoint that occurred before startTime
	checkpointTime := prevCheckpoint(obj.StartTime)
	return fillTimeAndGetCmd(&arg, usageInfluxDBTemplate, &checkpointTime, &checkpointTime)
}

func GetCloudletPoolUsageCommon(c echo.Context) error {
	rc := &InfluxDBContext{}
	regionRc := &RegionContext{}

	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	regionRc.username = claims.Username
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "usage/cloudletpool") {
		in := ormapi.RegionCloudletPoolUsage{}
		success, err := ReadConn(c, &in)
		if !success {
			return err
		}
		// Operator and cloudletpool name has to be specified
		if in.CloudletPool.Organization == "" || in.CloudletPool.Name == "" {
			return setReply(c, fmt.Errorf("CloudletPool details must be present"), nil)
		}
		rc.region = in.Region
		regionRc.region = in.Region

		cloudletpoolQuery := edgeproto.CloudletPool{Key: in.CloudletPool}
		// this also does an authorization check, so we dont have to
		cloudletPools, err := ShowCloudletPoolObj(ctx, regionRc, &cloudletpoolQuery)
		// since we specify name, should only have at most 1 result
		if err != nil || len(cloudletPools) != 1 {
			return setReply(c, fmt.Errorf("Unable to retrieve CloudletPool info"), nil)
		}

		cloudletList := []string{}
		for _, cloudlet := range cloudletPools[0].Cloudlets {
			cloudletList = append(cloudletList, cloudlet)
		}

		// check clusters
		eventCmd := cloudletPoolEventsQuery(&in, cloudletList, CLUSTER)
		checkpointCmd := cloudletPoolCheckpointsQuery(&in, cloudletList, CLUSTER)
		eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
		if err != nil {
			return setReply(c, fmt.Errorf("Error retrieving usage records: %v", err), nil)
		}
		clusterUsage, err := GetClusterUsage(eventResp, checkResp, in.StartTime, in.EndTime, in.Region)
		if err != nil {
			return setReply(c, fmt.Errorf("Error calculating usage records: %v", err), nil)
		}

		// check VM appinsts
		eventCmd = cloudletPoolEventsQuery(&in, cloudletList, APPINST)
		checkpointCmd = cloudletPoolCheckpointsQuery(&in, cloudletList, APPINST)
		eventResp, checkResp, err = GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
		if err != nil {
			return setReply(c, fmt.Errorf("Error retrieving usage records: %v", err), nil)
		}
		appUsage, err := GetAppUsage(eventResp, checkResp, in.StartTime, in.EndTime, in.Region)
		if err != nil {
			return setReply(c, fmt.Errorf("Error calculating usage records: %v", err), nil)
		}

		// sort it into cloudletPoolUsage struct
		usage, err := SortCloudletPoolUsage(&in, clusterUsage, appUsage, cloudletList)
		if err != nil {
			return setReply(c, fmt.Errorf("Error sorting records"), nil)
		}
		return setReply(c, nil, usage)

	} else if strings.HasSuffix(c.Path(), "usage/registerpool") {
		in := ormapi.RegionClusterInstUsage{}
		success, err := ReadConn(c, &in)
		if !success {
			return err
		}
		// Developer org name has to be specified
		if in.ClusterInst.Organization == "" {
			return setReply(c, fmt.Errorf("Cluster details must be present"), nil)
		}
		rc.region = in.Region
		org := in.ClusterInst.Organization

		eventCmd := ClusterUsageEventsQuery(&in)
		checkpointCmd := ClusterCheckpointsQuery(&in)

		// Check the developer org against who is logged in
		if err := authorized(ctx, rc.claims.Username, org, ResourceClusterAnalytics, ActionView); err != nil {
			return err
		}

		eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
		if err != nil {
			return err
		}
		usage, err := GetClusterUsage(eventResp, checkResp, in.StartTime, in.EndTime, in.Region)
		if err != nil {
			return err
		}
		payload := ormapi.StreamPayload{}
		payload.Data = &usage.Data
		WriteStream(c, &payload)
	} else {
		return setReply(c, echo.ErrNotFound, nil)
	}

	return nil
}
