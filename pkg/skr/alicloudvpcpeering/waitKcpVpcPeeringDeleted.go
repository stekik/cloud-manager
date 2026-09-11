package alicloudvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitKcpVpcPeeringDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.KcpVpcPeering == nil {
		logger.Info("KCP AlicloudVpcPeering is deleted")
		return nil, ctx
	}

	logger.Info("Waiting for KCP AlicloudVpcPeering to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
