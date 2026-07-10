package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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
	if instanceId == "" {
		return nil, ctx
	}

	info, err := state.client.DescribeInstance(ctx, instanceId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud r-kvstore cluster instance")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed loading AlicloudRedisCluster: %s", err),
		})
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error updating RedisCluster status after failed DescribeInstance",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.instance = info
	return nil, ctx
}
