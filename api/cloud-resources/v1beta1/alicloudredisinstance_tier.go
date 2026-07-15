package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// The tier letter+number encodes the underlying AliCloud r-kvstore standard
// instance class and read-only replica count:
//
//	S1-S5 → redis.master.*.cloud with ReadOnlyCount=0 (HA master+replica)
//	P1-P5 → redis.master.*.cloud with ReadOnlyCount=1 (HA master+replica + read-only replica)
//
// Cloud-disk (*.cloud) classes are used because they support all engine versions
// including 7.0; local-disk (*.default) classes only support up to 6.0.
//
// Both letter (S↔P) and number (1..5) are mutable via ModifyInstanceSpec; no
// recreation is required to switch between S and P at the same capacity.
//
// The tier→InstanceClass mapping lives in pkg/skr/alicloudredisinstance/util.go
// and can be updated without a CRD version bump. Availability of a given class
// in a specific region is validated at runtime via DescribeAvailableResource.
//
// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;P1;P2;P3;P4;P5
type AlicloudRedisTier string

const (
	// S — Standard HA, master + replica, no read-only replica.
	AlicloudRedisTierS1 AlicloudRedisTier = "S1"
	AlicloudRedisTierS2 AlicloudRedisTier = "S2"
	AlicloudRedisTierS3 AlicloudRedisTier = "S3"
	AlicloudRedisTierS4 AlicloudRedisTier = "S4"
	AlicloudRedisTierS5 AlicloudRedisTier = "S5"

	// P — Premium HA, master + replica + one read-only replica.
	AlicloudRedisTierP1 AlicloudRedisTier = "P1"
	AlicloudRedisTierP2 AlicloudRedisTier = "P2"
	AlicloudRedisTierP3 AlicloudRedisTier = "P3"
	AlicloudRedisTierP4 AlicloudRedisTier = "P4"
	AlicloudRedisTierP5 AlicloudRedisTier = "P5"
)
