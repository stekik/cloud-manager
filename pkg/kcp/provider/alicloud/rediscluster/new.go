package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		shared := st.(types.State)
		state, err := stateFactory.NewState(ctx, shared)
		if err != nil {
			err = fmt.Errorf("error creating new alicloud rediscluster state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActionsNoName(
			actions.AddCommonFinalizer(),
			loadRedis,

			// delete ================================================================================
			composed.If(composed.MarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					removeReadyCondition,
					deleteRedis,
					waitRedisDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),

			// create/update =========================================================================
			composed.If(composed.NotMarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					createRedis,
					waitRedisAvailable,
					modifyInstanceClass,
					waitRedisAvailable,
					modifyShardCount,
					waitRedisAvailable,
					modifyReplicasPerShard,
					waitRedisAvailable,
					updateStatus,
				),
			),

			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
