package rediscluster

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyParameters calls ModifyInstanceConfig when the desired parameters in
// the KCP spec diverge from the current config persisted on the cluster.
// The AliCloud API accepts the config as a JSON object string.
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
	if len(desired) == 0 {
		return nil, ctx
	}

	current := map[string]string{}
	if state.instance.Config != "" {
		if err := json.Unmarshal([]byte(state.instance.Config), &current); err != nil {
			current = nil
		}
	}

	if current != nil && reflect.DeepEqual(current, desired) {
		return nil, ctx
	}

	configBytes, err := json.Marshal(desired)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error marshalling AliCloud r-kvstore cluster parameters",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	if err := state.client.ModifyInstanceConfig(ctx, state.instance.InstanceId, string(configBytes)); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster config",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
