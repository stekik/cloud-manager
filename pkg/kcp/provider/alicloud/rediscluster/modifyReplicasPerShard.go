package rediscluster

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyReplicasPerShard issues ModifyInstanceSpec if the desired
// ReplicasPerShard drifts from the observed state. This is a separate action
// from modifyInstanceClass so the sequence is always:
//  1. modifyInstanceClass (InstanceClass + ReplicasPerShard together if class changes)
//  2. waitRedisAvailable
//  3. modifyShardCount
//  4. waitRedisAvailable
//  5. modifyReplicasPerShard (replica-only drift after shard changes settled)
//  6. waitRedisAvailable
//
// In practice most reconciles touch only one dimension, so steps 1-5 are no-ops.
func modifyReplicasPerShard(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desired := kcp.Spec.Instance.Alicloud.ReplicasPerShard
	// Proxy-based cluster classes encode replica topology in the class name;
	// ReadOnlyCount is always 0 and cannot be tuned independently.
	if alicloudclient.IsProxyClusterClass(state.instance.InstanceClass) {
		return nil, ctx
	}
	if desired == state.instance.ReadOnlyCount {
		return nil, ctx
	}

	opts := alicloudclient.ModifyInstanceSpecOptions{
		ReadOnlyCount: tea.Int32(desired),
	}
	if err := state.client.ModifyInstanceSpec(ctx, state.instance.InstanceId, opts); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster replicas per shard",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
