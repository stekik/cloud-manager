package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyShardCount grows or shrinks the cluster to match the desired ShardCount.
// Per issue #2012 design decision 8, ShardCount must be changed in a separate
// ModifyInstanceSpec call from InstanceClass changes.
func modifyShardCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desired := kcp.Spec.Instance.Alicloud.ShardCount
	current := state.instance.ShardCount

	if desired == current {
		return nil, ctx
	}

	var err error
	if desired > current {
		err = state.client.AddShardingNode(ctx, state.instance.InstanceId, desired)
	} else {
		err = state.client.DeleteShardingNode(ctx, state.instance.InstanceId, desired)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster shard count",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
