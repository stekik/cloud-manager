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

func loadRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance != nil {
		return nil, ctx
	}

	instanceId := state.ObjAsRedisCluster().Status.Id
	if instanceId != "" {
		info, err := state.client.DescribeInstance(ctx, instanceId)
		if err != nil {
			if alicloudclient.IsNotFoundErr(err) {
				// Instance already gone — treat as not found so deletion can proceed.
				return nil, ctx
			}
			logger.Error(err, "Error describing AliCloud r-kvstore cluster instance")
			meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRedisCluster,
				Message: fmt.Sprintf("Failed loading AlicloudRedisCluster: %s", err),
			})
			if updErr := state.UpdateObjStatus(ctx); updErr != nil {
				return composed.LogErrorAndReturn(updErr,
					"Error updating RedisCluster status after failed DescribeInstance",
					composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}
		state.instance = info
		return nil, ctx
	}

	// No Status.Id yet — search by name to recover a previously created instance
	// whose ID was not persisted (crash between CreateInstance and status write).
	info, err := state.client.DescribeInstanceByName(ctx, state.ObjAsRedisCluster().Name)
	if err != nil {
		logger.Error(err, "Error searching AliCloud r-kvstore cluster instance by name")
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
	}
	if info != nil {
		logger.Info("Recovered AliCloud r-kvstore cluster instance by name", "instanceId", info.InstanceId)
		state.ObjAsRedisCluster().Status.Id = info.InstanceId
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error persisting recovered AliCloud r-kvstore cluster instance ID",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		state.instance = info
	}
	return nil, ctx
}
