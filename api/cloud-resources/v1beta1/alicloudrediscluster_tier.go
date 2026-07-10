package v1beta1

// AlicloudRedisClusterTier defines the per-shard capacity tier for an
// AlicloudRedisCluster. The tier letter+number encodes the underlying AliCloud
// r-kvstore cloud-native cluster instance class:
//
//	C3 → redis.shard.large.ce     ( 4 GB per shard)
//	C4 → redis.shard.xlarge.ce    ( 8 GB per shard)
//	C5 → redis.shard.2xlarge.ce   (16 GB per shard)
//	C6 → redis.shard.4xlarge.ce   (32 GB per shard)
//	C7 → redis.shard.8xlarge.ce   (64 GB per shard)
//
// Total cluster capacity = per-shard memory × shardCount.
//
// C tiers start at C3 (4 GB/shard) to align with the minimum useful cluster
// size, matching the Azure Managed Redis convention where C tiers start above
// the smallest shard size.
//
// The tier→InstanceClass mapping lives in pkg/skr/alicloudrediscluster/util.go
// and can be updated without a CRD version bump. Availability of a given class
// in a specific region is validated at runtime via DescribeAvailableResource.
//
// +kubebuilder:validation:Enum=C3;C4;C5;C6;C7
type AlicloudRedisClusterTier string

const (
	AlicloudRedisClusterTierC3 AlicloudRedisClusterTier = "C3"
	AlicloudRedisClusterTierC4 AlicloudRedisClusterTier = "C4"
	AlicloudRedisClusterTierC5 AlicloudRedisClusterTier = "C5"
	AlicloudRedisClusterTierC6 AlicloudRedisClusterTier = "C6"
	AlicloudRedisClusterTierC7 AlicloudRedisClusterTier = "C7"
)
