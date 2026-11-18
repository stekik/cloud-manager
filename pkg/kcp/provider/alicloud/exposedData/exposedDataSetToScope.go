package exposedData

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func exposedDataSetToScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var list []string
	for _, gw := range state.natGateways {
		list = append(list, gw.IpAddresses...)
	}

	state.ObjAsScope().Status.ExposedData.NatGatewayIps = pie.Sort(pie.Unique(list))

	composed.LoggerFromCtx(ctx).
		WithValues("natGatewayIps", fmt.Sprintf("%v", state.ObjAsScope().Status.ExposedData.NatGatewayIps)).
		Info("Exposed Data AliCloud NAT Gateway IP addresses")

	return nil, ctx
}
