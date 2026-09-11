package alicloudvpcpeering

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	obj := state.ObjAsAlicloudVpcPeering()
	if obj.Status.Id != "" {
		return nil, ctx
	}

	id := uuid.NewString()

	if obj.Labels == nil {
		obj.Labels = map[string]string{}
	}
	obj.Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudVpcPeering with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AlicloudVpcPeering updated with ID label")

	obj.Status.Id = id

	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudVpcPeering status with ID", composed.StopWithRequeue, ctx)
	}

	logger.Info("SKR AlicloudVpcPeering updated with ID status")

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx
}
