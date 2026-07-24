package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyInstanceClass issues ModifyInstanceSpec if the desired InstanceClass
// drifts from the observed state. ShardCount changes are handled by modifyShardCount.
//
// For proxy-based classes (redis.logic.sharding.*) the class name encodes the
// shard count, so when a user changes shardCount, redisTierToInstanceClass
// produces a new class string with the updated shard count embedded. The
// pipeline calls modifyInstanceClass first and modifyShardCount second (each
// followed by waitRedisAvailable). AliCloud accepts the new class name with the
// updated shard count in the same call, so there is no ordering conflict.
func modifyInstanceClass(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desiredClass := kcp.Spec.Instance.Alicloud.InstanceClass
	desiredReplicas := kcp.Spec.Instance.Alicloud.ReplicasPerShard

	classDrift := desiredClass != "" && desiredClass != state.instance.InstanceClass
	// Proxy-based cluster classes (redis.logic.sharding.*) always have
	// ReadOnlyCount=0; non-proxy classes may have a separate replica count.
	replicasDrift := !alicloudclient.IsProxyClusterClass(desiredClass) &&
		desiredReplicas != state.instance.ReadOnlyCount
	if !classDrift && !replicasDrift {
		return nil, ctx
	}

	opts := alicloudclient.ModifyInstanceSpecOptions{}
	if classDrift {
		opts.InstanceClass = desiredClass
	}
	if replicasDrift {
		opts.ReadOnlyCount = &desiredReplicas
	}

	if err := state.client.ModifyInstanceSpec(ctx, state.instance.InstanceId, opts); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster instance class",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
