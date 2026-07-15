package alicloudredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	alicloudRedisInstance := state.ObjAsAlicloudRedisInstance()

	if state.KcpRedisInstance == nil {
		logger.Error(fmt.Errorf("kcpRedisInstance not found"), "KcpRedisInstance not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, ctx
	}

	instanceClass, readOnlyCount, err := redisTierToInstanceClassAndReadOnlyCount(alicloudRedisInstance.Spec.RedisTier)
	if err != nil {
		errMsg := "failed to map redisTier to instanceClass"
		logger.Error(err, errMsg, "redisTier", alicloudRedisInstance.Spec.RedisTier)
		alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(alicloudRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AlicloudRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AlicloudRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisInstance.Spec.Instance.Alicloud.InstanceClass = instanceClass
	state.KcpRedisInstance.Spec.Instance.Alicloud.ReadOnlyCount = readOnlyCount
	state.KcpRedisInstance.Spec.Instance.Alicloud.Parameters = alicloudRedisInstance.Spec.Parameters

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
