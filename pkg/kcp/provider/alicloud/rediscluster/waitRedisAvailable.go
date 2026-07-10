package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	switch state.instance.InstanceStatus {
	case alicloudclient.InstanceStatusNormal:
		return nil, ctx
	case alicloudclient.InstanceStatusCreating, alicloudclient.InstanceStatusChanging:
		state.instance = nil
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	default:
		return composed.LogErrorAndReturn(
			fmt.Errorf("unexpected AliCloud r-kvstore status %q", state.instance.InstanceStatus),
			"AliCloud r-kvstore cluster instance in unexpected state",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}
}
