package alicloudrediscluster

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	alicloudRedisCluster := state.ObjAsAlicloudRedisCluster()

	if state.KcpRedisCluster == nil {
		logger.Error(fmt.Errorf("kcpRedisCluster not found"), "KcpRedisCluster not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, ctx
	}

	instanceClass, err := redisTierToInstanceClass(alicloudRedisCluster.Spec.RedisTier, alicloudRedisCluster.Spec.ShardCount)
	if err != nil {
		errMsg := "failed to map redisTier to instanceClass"
		logger.Error(err, errMsg, "redisTier", alicloudRedisCluster.Spec.RedisTier)
		alicloudRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(alicloudRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AlicloudRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AlicloudRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisCluster.Spec.Instance.Alicloud.InstanceClass = instanceClass
	state.KcpRedisCluster.Spec.Instance.Alicloud.EngineVersion = alicloudRedisCluster.Spec.EngineVersion
	state.KcpRedisCluster.Spec.Instance.Alicloud.ShardCount = alicloudRedisCluster.Spec.ShardCount
	state.KcpRedisCluster.Spec.Instance.Alicloud.ReplicasPerShard = alicloudRedisCluster.Spec.ReplicasPerShard
	state.KcpRedisCluster.Spec.Instance.Alicloud.Parameters = alicloudRedisCluster.Spec.Parameters

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
