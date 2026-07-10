package redisinstance

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

// createRedis provisions a new r-kvstore instance if one does not yet exist.
// The password is expected to be supplied via the SKR AuthSecret (Phase 5);
// until Phase 5 lands, no password is passed and AliCloud generates one on
// its side (later reset via ResetAccountPassword when the SKR side wires up).
func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance != nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisInstance()
	if kcp.Spec.Instance.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("spec.instance.alicloud is nil"),
			"AliCloud redisinstance without alicloud provider spec",
			composed.StopAndForget, ctx)
	}
	if state.IpRange() == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("ipRange is nil"),
			"AliCloud redisinstance requires resolved IpRange",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	vSwitchId := ""
	for _, sn := range state.IpRange().Status.Subnets {
		if sn.Id != "" {
			vSwitchId = sn.Id
			break
		}
	}
	opts := alicloudclient.CreateInstanceOptions{
		InstanceName:  kcp.Name,
		InstanceClass: kcp.Spec.Instance.Alicloud.InstanceClass,
		EngineVersion: kcp.Spec.Instance.Alicloud.EngineVersion,
		VpcId:         state.IpRange().Status.VpcId,
		VSwitchId:     vSwitchId,
		ReadOnlyCount: kcp.Spec.Instance.Alicloud.ReadOnlyCount,
		Token:         string(kcp.UID),
	}

	instanceId, err := state.client.CreateInstance(ctx, opts)
	if err != nil {
		logger.Error(err, "Error creating AliCloud r-kvstore instance")
		meta.SetStatusCondition(kcp.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating AlicloudRedis: %s", err),
		})
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error updating RedisInstance status after failed CreateInstance",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	kcp.Status.Id = instanceId
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error persisting new AliCloud r-kvstore instance ID",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
