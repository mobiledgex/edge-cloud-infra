package main

import "github.com/mobiledgex/edge-cloud/edgeproto"
import "fmt"

func main() {
	app := edgeproto.AppInstKey{
		AppKey: edgeproto.AppKey{
			Organization: "apporg",
			Name:         "appname",
			Version:      "1.0",
		},
		ClusterInstKey: edgeproto.VirtualClusterInstKey{
			Organization: "clusterorg",
			ClusterKey: edgeproto.ClusterKey{
				Name: "clustername",
			},
			CloudletKey: edgeproto.CloudletKey{
				Name:         "cloudletname",
				Organization: "cloudorg",
			},
		},
	}
	fmt.Println(app.String())
}
