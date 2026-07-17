package rediscluster

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyInstanceClass issues ModifyInstanceSpec if the desired InstanceClass
// or ReplicasPerShard drifts from the observed state. Per issue #2012 design
// decision 8, this must NOT change ShardCount in the same call.
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
	// Proxy-based cluster classes (redis.logic.sharding.*) encode replicas in
	// the class name; ReadOnlyCount is always 0 on those instances and cannot
	// be changed independently. Only check replicasDrift for standard classes.
	replicasDrift := !alicloudclient.IsProxyClusterClass(desiredClass) &&
		desiredReplicas != state.instance.ReadOnlyCount
	if !classDrift && !replicasDrift {
		return nil, ctx
	}

	opts := alicloudclient.ModifyInstanceSpecOptions{
		ReadOnlyCount: tea.Int32(desiredReplicas),
	}
	if classDrift {
		opts.InstanceClass = desiredClass
	}

	if err := state.client.ModifyInstanceSpec(ctx, state.instance.InstanceId, opts); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster instance class/replicas",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
