package mock

import (
	"context"

	rediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	redisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
)

// redisInstanceClientView adapts redisStore to redisinstance.Client.
type redisInstanceClientView struct{ *redisStore }

var _ redisinstance.Client = (*redisInstanceClientView)(nil)

func (c *redisInstanceClientView) CreateInstance(ctx context.Context, opts redisinstance.CreateInstanceOptions) (string, error) {
	return c.createInstance(ctx, opts)
}

func (c *redisInstanceClientView) DescribeInstance(ctx context.Context, instanceId string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstance(ctx, instanceId)
}

func (c *redisInstanceClientView) ModifyInstanceSpec(ctx context.Context, instanceId string, opts redisinstance.ModifyInstanceSpecOptions) error {
	return c.modifyInstanceSpec(ctx, instanceId, opts)
}

func (c *redisInstanceClientView) DeleteInstance(ctx context.Context, instanceId string) error {
	return c.deleteInstance(ctx, instanceId)
}

func (c *redisInstanceClientView) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	return c.resetAccountPassword(ctx, instanceId, accountName, password)
}

// redisClusterClientView adapts redisStore to rediscluster.Client.
type redisClusterClientView struct{ *redisStore }

var _ rediscluster.Client = (*redisClusterClientView)(nil)

func (c *redisClusterClientView) CreateInstance(ctx context.Context, opts redisinstance.CreateInstanceOptions) (string, error) {
	return c.createInstance(ctx, opts)
}

func (c *redisClusterClientView) DescribeInstance(ctx context.Context, instanceId string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstance(ctx, instanceId)
}

func (c *redisClusterClientView) ModifyInstanceSpec(ctx context.Context, instanceId string, opts redisinstance.ModifyInstanceSpecOptions) error {
	return c.modifyInstanceSpec(ctx, instanceId, opts)
}

func (c *redisClusterClientView) DeleteInstance(ctx context.Context, instanceId string) error {
	return c.deleteInstance(ctx, instanceId)
}

func (c *redisClusterClientView) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	return c.resetAccountPassword(ctx, instanceId, accountName, password)
}

func (c *redisClusterClientView) AddShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	return c.addShardingNode(ctx, instanceId, targetShardCount)
}

func (c *redisClusterClientView) DeleteShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	return c.deleteShardingNode(ctx, instanceId, targetShardCount)
}
