package main

import (
	"github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator"
	"github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg"

	"github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetOperatorApiGw(operatorName string) (operator.OperatorApiGw, error) {
	log.DebugLog(log.DebugLevelMexos, "GetOperatorApiGw", "operatorName", operatorName)

	switch operatorName {
	case "tdg":
		fallthrough
	case "TDG":
		return &tdg.OperatorApiGw{}, nil
	}
	return &defaultoperator.OperatorApiGw{}, nil
}
