package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance == nil {
		return nil, ctx
	}
	if state.instance.InstanceStatus == alicloudclient.InstanceStatusReleased {
		return nil, ctx
	}
	if state.instance.InstanceStatus == alicloudclient.InstanceStatusCreating {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if err := state.client.DeleteInstance(ctx, state.instance.InstanceId); err != nil {
		logger.Error(err, "Error deleting AliCloud r-kvstore cluster instance")
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
