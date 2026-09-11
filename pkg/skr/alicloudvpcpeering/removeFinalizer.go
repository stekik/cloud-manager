package alicloudvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.KcpVpcPeering != nil {
		return nil, ctx
	}

	// Wait until the KCP remote Network is fully gone before removing the SKR finalizer.
	// deleteKcpRemoteNetwork only issues a Delete() call; the Network object may still
	// be terminating on the next cycle. Removing the SKR finalizer while the Network
	// object is still present would orphan it.
	if state.RemoteNetwork != nil {
		return nil, ctx
	}

	logger.Info("Removing AlicloudVpcPeering finalizer")

	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AlicloudVpcPeering after finalizer remove", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, ctx
}
