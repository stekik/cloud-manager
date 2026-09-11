package alicloudvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpVpcPeering == nil {
		return nil, ctx
	}

	obj := state.ObjAsAlicloudVpcPeering()

	changed := false

	if composed.SyncConditions(obj, *state.KcpVpcPeering.Conditions()...) {
		changed = true
	}

	if obj.Status.State != state.KcpVpcPeering.Status.State {
		changed = true
	}

	if changed {
		obj.Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(*state.KcpVpcPeering.Conditions()...).
			ErrorLogMessage("Error updating SKR AlicloudVpcPeering status").
			SuccessLogMsg("Updated SKR AlicloudVpcPeering status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
