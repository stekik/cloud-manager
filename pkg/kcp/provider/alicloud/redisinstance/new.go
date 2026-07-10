package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

// New returns the composed.Action for the AliCloud provider case in the
// shared pkg/kcp/redisinstance/reconciler.go switch. It follows the same
// shape as the AWS/Azure/GCP counterparts:
//
//  1. build a provider-specific State from the shared types.State,
//  2. run loadRedis,
//  3. branch on MarkedForDeletionPredicate between create/update and delete
//     pipelines (design: issue #2012 SKR pipeline note).
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		shared := st.(types.State)
		state, err := stateFactory.NewState(ctx, shared)
		if err != nil {
			err = fmt.Errorf("error creating new alicloud redisinstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"alicloudRedisInstance",
			actions.AddCommonFinalizer(),
			loadRedis,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"alicloudRedisInstance-create",
					createRedis,
					waitRedisAvailable,
					modifyInstanceClass,
					updateStatus,
				),
				composed.ComposeActions(
					"alicloudRedisInstance-delete",
					removeReadyCondition,
					deleteRedis,
					waitRedisDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
