package iprange

import (
	"context"
	"slices"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ipclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func rangeDisassociateVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	cidr := state.ObjAsIpRange().Status.Cidr
	if cidr == "" {
		return nil, ctx
	}

	if !slices.Contains(state.secondaryCidrBlocks, cidr) {
		return nil, ctx
	}

	// If we already set the CidrInUse error on a previous reconcile, don't call the
	// cloud API again — just wait. Status.State is reliably cache-consistent even on
	// watch-event-driven re-reconciles because the informer reflects it after the
	// first successful patch.
	if state.ObjAsIpRange().Status.State == cloudcontrolv1beta1.StateError {
		existing := meta.FindStatusCondition(state.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		if existing != nil && existing.Reason == cloudcontrolv1beta1.ReasonFailedDisassociatingVpcAddressSpace {
			return composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx
		}
	}

	logger.Info("Disassociating secondary CIDR block from AliCloud VPC", "vpcId", state.vpcId, "cidr", cidr)

	err := state.client.UnassociateVpcCidrBlock(ctx, state.vpcId, cidr)
	if err != nil {
		if ipclient.IsCidrInUseErr(err) {
			logger.Error(err, "CIDR block still in use by a vSwitch — manual vSwitch deletion required before IpRange can be deleted",
				"vpcId", state.vpcId, "cidr", cidr)

			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeError,
					Status:             metav1.ConditionTrue,
					ObservedGeneration: state.ObjAsIpRange().Generation,
					Reason:             cloudcontrolv1beta1.ReasonFailedDisassociatingVpcAddressSpace,
					Message:            "CIDR block is still in use by a vSwitch; delete the vSwitch in the cloud console to unblock deletion",
				}).
				ErrorLogMessage("Error patching AliCloud KCP IpRange status after CidrInUse error").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, state)
		}
		logger.Error(err, "Error disassociating secondary CIDR block from AliCloud VPC")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
