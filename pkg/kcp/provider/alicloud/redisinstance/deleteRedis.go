package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// deleteRedis issues DeleteInstance if the r-kvstore instance still exists.
// PostPaid-only is a design constraint (issue #2012 decision 5); PrePaid
// deletion is not supported.
func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance == nil {
		return nil, ctx
	}
	if state.instance.InstanceStatus == alicloudclient.InstanceStatusReleased {
		return nil, ctx
	}
	// AliCloud rejects DeleteInstance while the instance is still being created.
	// Wait for it to reach Normal before issuing the delete.
	if state.instance.InstanceStatus == alicloudclient.InstanceStatusCreating {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
	}

	if err := state.client.DeleteInstance(ctx, state.instance.InstanceId); err != nil {
		logger.Error(err, "Error deleting AliCloud r-kvstore instance")
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
	}
	// Force re-load next reconcile so waitRedisDeleted observes the change.
	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
