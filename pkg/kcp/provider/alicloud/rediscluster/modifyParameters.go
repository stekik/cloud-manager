package rediscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyParameters calls ModifyInstanceConfig when the desired parameters in
// the KCP spec diverge from the current config persisted on the cluster.
// The AliCloud API accepts the config as a JSON object string.
// An empty desired map sends "{}" to clear all custom parameters.
func modifyParameters(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desired := kcp.Spec.Instance.Alicloud.Parameters

	// Parse current config from instance. AliCloud returns the full config object
	// including system defaults. To avoid a perpetual modify loop when desired is
	// empty (or a subset of system defaults), project current down to only the keys
	// that appear in desired before comparing.
	currentFull := map[string]string{}
	if state.instance.Config != "" {
		raw := map[string]interface{}{}
		if err := json.Unmarshal([]byte(state.instance.Config), &raw); err == nil {
			for k, v := range raw {
				currentFull[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	current := map[string]string{}
	for k := range desired {
		if v, ok := currentFull[k]; ok {
			current[k] = v
		}
	}

	if maps.Equal(desired, current) {
		return nil, ctx
	}

	var configStr string
	if len(desired) == 0 {
		configStr = "{}"
	} else {
		configBytes, err := json.Marshal(desired)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error marshalling AliCloud r-kvstore cluster parameters",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
		}
		configStr = string(configBytes)
	}

	if err := state.client.ModifyInstanceConfig(ctx, state.instance.InstanceId, configStr); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster config",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
