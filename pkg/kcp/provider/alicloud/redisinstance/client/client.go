// Package client wraps the AliCloud r-kvstore SDK for use by the KCP AliCloud
// Redis reconcilers. It exposes a narrow, provider-neutral Client interface
// covering the standard (non-sharded) HA instance lifecycle.
//
// The interface intentionally mirrors the operations described in issue #2012:
//
//   - CreateInstance          → r-kvstore CreateInstance
//   - DescribeInstance        → r-kvstore DescribeInstanceAttribute
//   - ModifyInstanceSpec      → r-kvstore ModifyInstanceSpec
//   - DeleteInstance          → r-kvstore DeleteInstance
//   - ResetAccountPassword    → r-kvstore ResetAccountPassword
//
// Cluster-specific operations (AddShardingNode/DeleteShardingNode) live in the
// sibling rediscluster/client package which embeds this interface.
package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rkvstore "github.com/alibabacloud-go/r-kvstore-20150101/v7/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Instance status values used by the AliCloud r-kvstore API. Only a subset is
// consumed by the reconcilers; the constants are documented in issue #2012.
const (
	InstanceStatusCreating = "Creating"
	InstanceStatusNormal   = "Normal"
	InstanceStatusChanging = "Changing"
	InstanceStatusInactive = "Inactive"
	InstanceStatusReleased = "Released"
)

// NetworkType is always VPC for cloud-manager (public endpoints are not
// supported in v1 per design decision).
const NetworkType = "VPC"

// ChargeType is always PostPaid for cloud-manager (PrePaid deletion is not
// supported via API, deferred).
const ChargeType = "PostPaid"

// DefaultPort is the r-kvstore default listening port.
const DefaultPort int64 = 6379

// InstanceInfo is a provider-neutral snapshot of the state returned by
// DescribeInstanceAttribute.
type InstanceInfo struct {
	InstanceId       string
	InstanceStatus   string
	InstanceClass    string
	ArchitectureType string
	NodeType         string
	NetworkType      string
	VpcId            string
	VSwitchId        string
	EngineVersion    string
	Capacity         int64
	Port             int64
	ConnectionDomain string
	ChargeType       string
	ShardCount       int32
	ReadOnlyCount    int32
	Config           string
}

// CreateInstanceOptions are the arguments accepted by CreateInstance.
//
// ZoneId is intentionally omitted from this struct: the AliCloud API derives
// it automatically from VSwitchId (design decision 7 in issue #2012).
type CreateInstanceOptions struct {
	InstanceName  string
	InstanceClass string
	EngineVersion string
	VpcId         string
	VSwitchId     string
	Password      string
	Port          int64
	// ShardCount is only meaningful for cluster classes (redis.shard.*.ce).
	// Leave zero for standard instances.
	ShardCount int32
	// ReadOnlyCount encodes the S/P tier for standard instances (0/1).
	ReadOnlyCount int32
	// Token is an idempotency token; if set, repeated calls with the same
	// token return the same instance without creating a duplicate.
	Token string
}

// ModifyInstanceSpecOptions covers the ModifyInstanceSpec request surface
// used by the cloud-manager reconcilers. Fields set to zero value are omitted
// from the request.
type ModifyInstanceSpecOptions struct {
	InstanceClass string
	ShardCount    int32
	ReadOnlyCount int32
}

// Client is the AliCloud r-kvstore standard-instance client contract.
type Client interface {
	CreateInstance(ctx context.Context, opts CreateInstanceOptions) (instanceId string, err error)
	DescribeInstance(ctx context.Context, instanceId string) (*InstanceInfo, error)
	ModifyInstanceSpec(ctx context.Context, instanceId string, opts ModifyInstanceSpecOptions) error
	DeleteInstance(ctx context.Context, instanceId string) error
	ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error
}

// ClientProvider is the standard cloud-manager credential/region-scoped
// constructor signature used across all AliCloud client packages.
type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

// NewClientProvider returns a ClientProvider that constructs a real SDK-backed
// AliCloud r-kvstore client.
func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		config := &openapi.Config{
			AccessKeyId:     new(accessKeyId),
			AccessKeySecret: new(accessKeySecret),
			RegionId:        new(region),
		}
		config.Endpoint = new(fmt.Sprintf("r-kvstore.%s.aliyuncs.com", region))

		rc, err := rkvstore.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud r-kvstore client: %w", err)
		}

		return &alicloudRedisClient{
			c:      rc,
			region: region,
		}, nil
	}
}

type alicloudRedisClient struct {
	c      *rkvstore.Client
	region string
}

// CreateInstance provisions a new r-kvstore instance. The instance
// architecture (standard vs cluster) is entirely determined by
// opts.InstanceClass:
//   - redis.master.*.default → standard HA
//   - redis.shard.*.ce       → cloud-native cluster (requires ShardCount)
func (c *alicloudRedisClient) CreateInstance(ctx context.Context, opts CreateInstanceOptions) (string, error) {
	req := &rkvstore.CreateInstanceRequest{
		RegionId:      new(c.region),
		InstanceName:  new(opts.InstanceName),
		InstanceClass: new(opts.InstanceClass),
		EngineVersion: new(opts.EngineVersion),
		VpcId:         new(opts.VpcId),
		VSwitchId:     new(opts.VSwitchId),
		NetworkType:   new(NetworkType),
		ChargeType:    new(ChargeType),
	}
	if opts.Password != "" {
		req.Password = new(opts.Password)
	}
	if opts.Port > 0 {
		portStr := fmt.Sprintf("%d", opts.Port)
		req.Port = new(portStr)
	}
	if opts.ShardCount > 0 {
		req.ShardCount = new(opts.ShardCount)
	}
	if opts.ReadOnlyCount > 0 {
		req.ReadOnlyCount = new(opts.ReadOnlyCount)
	}
	if opts.Token != "" {
		req.Token = new(opts.Token)
	}

	resp, err := c.c.CreateInstance(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud r-kvstore instance: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("empty response from alicloud r-kvstore CreateInstance")
	}
	return tea.StringValue(resp.Body.InstanceId), nil
}

// DescribeInstance loads the current state of an instance. Returns (nil, nil)
// when the instance is not found (caller should treat as "does not exist").
func (c *alicloudRedisClient) DescribeInstance(ctx context.Context, instanceId string) (*InstanceInfo, error) {
	req := &rkvstore.DescribeInstanceAttributeRequest{
		InstanceId: new(instanceId),
	}
	resp, err := c.c.DescribeInstanceAttribute(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	if resp == nil || resp.Body == nil || resp.Body.Instances == nil ||
		len(resp.Body.Instances.DBInstanceAttribute) == 0 {
		return nil, nil
	}
	a := resp.Body.Instances.DBInstanceAttribute[0]
	return &InstanceInfo{
		InstanceId:       tea.StringValue(a.InstanceId),
		InstanceStatus:   tea.StringValue(a.InstanceStatus),
		InstanceClass:    tea.StringValue(a.InstanceClass),
		ArchitectureType: tea.StringValue(a.ArchitectureType),
		NodeType:         tea.StringValue(a.NodeType),
		NetworkType:      tea.StringValue(a.NetworkType),
		VpcId:            tea.StringValue(a.VpcId),
		VSwitchId:        tea.StringValue(a.VSwitchId),
		EngineVersion:    tea.StringValue(a.EngineVersion),
		Capacity:         tea.Int64Value(a.Capacity),
		Port:             tea.Int64Value(a.Port),
		ConnectionDomain: tea.StringValue(a.ConnectionDomain),
		ChargeType:       tea.StringValue(a.ChargeType),
		ShardCount:       tea.Int32Value(a.ShardCount),
		ReadOnlyCount:    tea.Int32Value(a.ReadOnlyCount),
		Config:           tea.StringValue(a.Config),
	}, nil
}

// ModifyInstanceSpec scales an existing instance. Per issue #2012 design
// decision 8, callers must not combine InstanceClass changes with ShardCount
// changes in the same request; the cluster reconciler splits these into two
// separate calls.
func (c *alicloudRedisClient) ModifyInstanceSpec(ctx context.Context, instanceId string, opts ModifyInstanceSpecOptions) error {
	req := &rkvstore.ModifyInstanceSpecRequest{
		InstanceId: new(instanceId),
	}
	if opts.InstanceClass != "" {
		req.InstanceClass = new(opts.InstanceClass)
	}
	if opts.ShardCount > 0 {
		req.ShardCount = new(opts.ShardCount)
	}
	// ReadOnlyCount = 0 is a valid target (transition P→S), so we always send
	// it when the caller explicitly opts in via a non-negative value. To
	// distinguish "unset" from "zero" callers should call this method only
	// when they know a change is required.
	req.ReadOnlyCount = new(opts.ReadOnlyCount)

	if _, err := c.c.ModifyInstanceSpec(req); err != nil {
		return fmt.Errorf("error modifying alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	return nil
}

// DeleteInstance releases a PostPaid r-kvstore instance. PrePaid instances
// cannot be deleted via API (design decision 5) and are not supported in v1.
func (c *alicloudRedisClient) DeleteInstance(ctx context.Context, instanceId string) error {
	req := &rkvstore.DeleteInstanceRequest{
		InstanceId: new(instanceId),
	}
	if _, err := c.c.DeleteInstance(req); err != nil {
		return fmt.Errorf("error deleting alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	return nil
}

// ResetAccountPassword resets the password of a named account on an existing
// instance. Used only in edge-cases where the pre-creation password is lost
// (e.g. AuthSecret regeneration). Cloud-manager normally sets the initial
// password at CreateInstance time (design decision 6).
func (c *alicloudRedisClient) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	req := &rkvstore.ResetAccountPasswordRequest{
		InstanceId:      new(instanceId),
		AccountName:     new(accountName),
		AccountPassword: new(password),
	}
	if _, err := c.c.ResetAccountPassword(req); err != nil {
		return fmt.Errorf("error resetting password for alicloud r-kvstore instance %s account %s: %w", instanceId, accountName, err)
	}
	return nil
}
