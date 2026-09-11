package alicloudvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.RemoteNetwork == nil {
		return nil, ctx
	}

	if composed.IsMarkedForDeletion(state.RemoteNetwork) {
		return nil, ctx
	}

	logger.Info("Deleting KCP remote Network")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.RemoteNetwork)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP remote Network", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
