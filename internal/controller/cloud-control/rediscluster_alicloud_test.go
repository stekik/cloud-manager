package cloudcontrol

import (
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP AliCloud RedisCluster", func() {

	It("Scenario: KCP AliCloud RedisCluster is created and deleted", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "e5f6a7b8-c9d0-1234-efab-345678901234"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "f6a7b8c9-d0e1-2345-fabc-456789012345"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		instanceClass := "redis.shard.large.ce"
		engineVersion := "7.0"
		shardCount := int32(3)

		By("When RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass(instanceClass),
					WithKcpAlicloudRedisEngineVersion(engineVersion),
					WithKcpAlicloudRedisClusterShardCount(shardCount),
				).Should(Succeed(), "failed creating RedisCluster")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("Then AliCloud Redis cluster is created and gets an ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected RedisCluster to get status.id")
		})

		By("When AliCloud Redis transitions to Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed(), "expected RedisCluster to reach Ready state")
		})

		By("And Then RedisCluster has .status.discoveryEndpoint set", func() {
			Expect(len(redisCluster.Status.DiscoveryEndpoint) > 0).To(BeTrue())
		})

		By("And Then RedisCluster has .status.authString set", func() {
			Expect(len(redisCluster.Status.AuthString) > 0).To(BeTrue())
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster to be deleted")
		})
	})

	It("Scenario: KCP AliCloud RedisCluster shard count is scaled up", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "a7b8c9d0-e1f2-3456-abcd-567890123456"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "b8c9d0e1-f2a3-4567-bcde-678901234567"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-scale"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(3),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisCluster gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("And Given AliCloud Redis is Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("And Given RedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		By("When shard count is increased to 6", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.ShardCount = 6
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then AliCloud transitions to Changing and back to Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisCluster remains Ready after scale-up", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		By("And Then RedisCluster status.shardCount is 6", func() {
			Expect(redisCluster.Status.ShardCount).To(Equal(int32(6)))
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})
})
