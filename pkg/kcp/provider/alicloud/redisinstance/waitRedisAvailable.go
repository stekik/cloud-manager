package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitRedisAvailable polls the instance status until it reaches Normal. While
// Creating or Changing, the reconciler requeues. Any other terminal state is
// surfaced as an error condition.
func waitRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	switch state.instance.InstanceStatus {
	case alicloudclient.InstanceStatusNormal:
		return nil, ctx
	case alicloudclient.InstanceStatusCreating, alicloudclient.InstanceStatusChanging:
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	default:
		return composed.LogErrorAndReturn(
			fmt.Errorf("unexpected AliCloud r-kvstore status %q", state.instance.InstanceStatus),
			"AliCloud r-kvstore instance in unexpected state",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}
}
