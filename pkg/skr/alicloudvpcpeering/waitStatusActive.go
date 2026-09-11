package alicloudvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitStatusActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAlicloudVpcPeering()

	// If the KCP VpcPeering reached a terminal error state, stop waiting — the user
	// must correct the spec or delete the resource; retrying will not help.
	if meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeError) &&
		(obj.Status.State == string(cloudcontrolv1beta1.StateError) ||
			obj.Status.State == string(cloudcontrolv1beta1.StateWarning)) {
		return composed.StopAndForget, ctx
	}

	if !meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) ||
		obj.Status.State != string(cloudcontrolv1beta1.StateReady) {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	return nil, ctx
}
