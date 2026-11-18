package exposedData

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func natGatewayLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsScope().Spec.Scope.Alicloud == nil {
		return nil, ctx
	}
	vpcId := state.ObjAsScope().Spec.Scope.Alicloud.Network.VPC.Id
	if vpcId == "" {
		composed.LoggerFromCtx(ctx).Info("AliCloud Scope VPC ID not set, skipping NAT gateway load")
		return nil, ctx
	}

	gateways, err := state.client.DescribeNatGateways(ctx, vpcId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading AliCloud NAT gateways", composed.StopWithRequeue, ctx)
	}

	state.natGateways = gateways

	return nil, ctx
}
