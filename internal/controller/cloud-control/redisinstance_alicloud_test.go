package cloudcontrol

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP AliCloud RedisInstance", func() {

	It("Scenario: KCP AliCloud RedisInstance is created and deleted", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "b2c3d4e5-f6a7-8901-bcde-f12345678901"
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		instanceClass := "redis.master.small.default"
		engineVersion := "7.0"

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-alicloud-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAlicloud(),
					WithKcpAlicloudRedisInstanceClass(instanceClass),
					WithKcpAlicloudRedisEngineVersion(engineVersion),
				).Should(Succeed(), "failed creating RedisInstance")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("Then AliCloud Redis is created in Creating status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected RedisInstance to get status.id")
		})

		By("When AliCloud Redis transitions to Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed(), "expected RedisInstance to reach Ready state")
		})

		By("And Then RedisInstance has .status.primaryEndpoint set", func() {
			Expect(len(redisInstance.Status.PrimaryEndpoint) > 0).To(BeTrue())
		})

		By("And Then RedisInstance has .status.authString set", func() {
			Expect(len(redisInstance.Status.AuthString) > 0).To(BeTrue())
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance to be deleted")
		})
	})

	It("Scenario: KCP AliCloud RedisInstance load error is surfaced", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "c3d4e5f6-a7b8-9012-cdef-123456789012"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "d4e5f6a7-b8c9-0123-defg-234567890123"
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-alicloud-redis-error"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAlicloud(),
					WithKcpAlicloudRedisInstanceClass("redis.master.small.default"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
				).Should(Succeed(), "failed creating RedisInstance")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisInstance gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("When AliCloud returns an error on describe", func() {
			alicloudMock.SetRedisInstanceError(redisInstance.Status.Id, fmt.Errorf("simulated AliCloud API failure"))
		})

		By("Then RedisInstance has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError),
				).Should(Succeed(), "expected RedisInstance to surface error condition")
		})

		By("When error is cleared", func() {
			alicloudMock.SetRedisInstanceError(redisInstance.Status.Id, nil)
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisInstance recovers to Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		// cleanup
		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})
	})
})
