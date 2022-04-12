// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg"
	"github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	"github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetOperatorApiGw(ctx context.Context, operatorName string) (operator.OperatorApiGw, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetOperatorApiGw", "operatorName", operatorName)

	var outApiGw operator.OperatorApiGw
	switch operatorName {
	case "tdg":
		fallthrough
	case "TDG":
		outApiGw = &tdg.OperatorApiGw{}
	default:
		outApiGw = &defaultoperator.OperatorApiGw{}
	}
	return outApiGw, nil
}
