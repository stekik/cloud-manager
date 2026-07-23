/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

// RedisInstanceAlicloud describes an AliCloud r-kvstore standard (non-sharded)
// HA instance in the KCP RedisInstance CRD.
//
// InstanceClass is the AliCloud r-kvstore instance class string
// (e.g. "tair.rdb.4g"). It is resolved by the SKR reconciler from the
// SKR-side redisTier abstraction.
//
// ReadOnlyCount encodes the service tier: 0 = S tier (no read-only replica),
// 1 = P tier (one read-only replica). Both InstanceClass and ReadOnlyCount are
// mutable via ModifyInstanceSpec, so no CEL immutability rules are attached.
type RedisInstanceAlicloud struct {
	// +kubebuilder:validation:Required
	InstanceClass string `json:"instanceClass"`

	// EngineVersion is set at creation and is immutable. Defaults to "5.0" for
	// standard instances (broad class compatibility); cluster instances default
	// to "7.0" (see RedisClusterAlicloud.EngineVersion).
	// +optional
	// +kubebuilder:default="5.0"
	// +kubebuilder:validation:Enum="5.0";"6.0";"7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf),message="engineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// ReadOnlyCount: 0 = S tier (no read replica), 1 = P tier (one read replica).
	// +optional
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	ReadOnlyCount int32 `json:"readOnlyCount"`

	// Parameters are passed to the AliCloud instance as runtime configuration.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

// RedisClusterAlicloud describes an AliCloud r-kvstore sharded cloud-native
// cluster instance in the KCP RedisCluster CRD.
//
// InstanceClass is the per-shard AliCloud r-kvstore instance class string
// (e.g. "redis.logic.sharding.4g.3db.0rodb.4proxy.default"). It is resolved
// by the SKR reconciler from the SKR-side redisTier abstraction using a
// proxy-based sharding class.
//
// AliCloud does not support changing InstanceClass and ShardCount in the same
// ModifyInstanceSpec call. The KCP cluster pipeline uses two separate modify
// actions, each followed by a wait for the instance to become available. See
// design decision 8 in issue #2012.
type RedisClusterAlicloud struct {
	// +kubebuilder:validation:Required
	InstanceClass string `json:"instanceClass"`

	// EngineVersion is set at creation and is immutable. Defaults to "7.0" for
	// cluster instances; standard instances default to "5.0" (see
	// RedisInstanceAlicloud.EngineVersion).
	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:Enum="5.0";"6.0";"7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf),message="engineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// ShardCount is the number of data shards in the cluster.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=32
	ShardCount int32 `json:"shardCount"`

	// ReplicasPerShard: 0 = no replica per shard, 1 = HA per shard.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	ReplicasPerShard int32 `json:"replicasPerShard"`

	// Parameters are passed to the AliCloud instance as runtime configuration.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}
