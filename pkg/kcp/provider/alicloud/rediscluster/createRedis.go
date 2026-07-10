package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance != nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("spec.instance.alicloud is nil"),
			"AliCloud rediscluster without alicloud provider spec",
			composed.StopAndForget, ctx)
	}
	if state.IpRange() == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("ipRange is nil"),
			"AliCloud rediscluster requires resolved IpRange",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	vSwitchId := ""
	for _, sn := range state.IpRange().Status.Subnets {
		if sn.Id != "" {
			vSwitchId = sn.Id
			break
		}
	}
	if vSwitchId == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("no vSwitch found in IpRange subnets"),
			"AliCloud rediscluster IpRange has no vSwitch",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Generate password before CreateInstance — AliCloud never returns it after.
	// If AuthString is already set (idempotent retry), reuse it.
	password := kcp.Status.AuthString
	if password == "" {
		password = util.RandomString(32)
	}

	opts := alicloudclient.CreateInstanceOptions{
		InstanceName:  kcp.Name,
		InstanceClass: kcp.Spec.Instance.Alicloud.InstanceClass,
		EngineVersion: kcp.Spec.Instance.Alicloud.EngineVersion,
		VpcId:         state.IpRange().Status.VpcId,
		VSwitchId:     vSwitchId,
		Password:      password,
		ShardCount:    kcp.Spec.Instance.Alicloud.ShardCount,
		ReadOnlyCount: kcp.Spec.Instance.Alicloud.ReplicasPerShard,
		Token:         string(kcp.UID),
	}

	instanceId, err := state.client.CreateInstance(ctx, opts)
	if err != nil {
		logger.Error(err, "Error creating AliCloud r-kvstore cluster instance")
		meta.SetStatusCondition(kcp.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRedisCluster,
			Message: fmt.Sprintf("Failed creating AlicloudRedisCluster: %s", err),
		})
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error updating RedisCluster status after failed CreateInstance",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	kcp.Status.Id = instanceId
	kcp.Status.AuthString = password
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error persisting new AliCloud r-kvstore cluster instance ID and auth string",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
