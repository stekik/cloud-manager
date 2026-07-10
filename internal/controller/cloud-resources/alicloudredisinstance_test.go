package cloudresources

import (
	"github.com/google/uuid"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Feature: SKR AlicloudRedisInstance", func() {

	It("Scenario: SKR AlicloudRedisInstance is created and deleted", func() {

		skrIpRangeName := uuid.NewString()
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := uuid.NewString()

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(skrIpRangeName)
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).Should(Succeed())
		})

		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		alicloudRedisInstanceName := uuid.NewString()
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: alicloudRedisInstanceName}))
		alicloudRedisInstance := &cloudresourcesv1beta1.AlicloudRedisInstance{}

		const authSecretName = "alicloud-redis-auth-secret"
		authSecretLabels := map[string]string{"env": "test"}
		authSecretAnnotations := map[string]string{"note": "alicloud"}
		extraData := map[string]string{
			"endpoint": "{{.host}}:{{.port}}",
		}

		By("When AlicloudRedisInstance is created", func() {
			Eventually(CreateAlicloudRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					WithName(alicloudRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisInstanceRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
					WithAlicloudRedisInstanceEngineVersion("7.0"),
					WithAlicloudRedisInstanceAuthSecretName(authSecretName),
					WithAlicloudRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAlicloudRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
					WithAlicloudRedisInstanceAuthSecretExtraData(extraData),
				).Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected SKR AlicloudRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(WithName(alicloudRedisInstance.Status.Id)),
				).Should(Succeed())

			By("And it has KCP labels")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(alicloudRedisInstance.Name))
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(alicloudRedisInstance.Namespace))

			By("And spec.instance.alicloud matches SKR tier")
			Expect(kcpRedisInstance.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.master.small.default"))
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.ReadOnlyCount).To(Equal(int32(0)))
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.EngineVersion).To(Equal("7.0"))
		})

		kcpPrimaryEndpoint := "r-test123.redis.rds.aliyuncs.com:6379"
		kcpAuthString := uuid.NewString()

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpPrimaryEndpoint),
					WithRedisInstanceAuthString(kcpAuthString),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Then SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret,
					NewObjActions(
						WithName(authSecretName),
						WithNamespace(alicloudRedisInstance.Namespace),
					),
					HavingLabelKeys(
						util.WellKnownK8sLabelComponent,
						util.WellKnownK8sLabelPartOf,
						util.WellKnownK8sLabelManagedBy,
					),
					HavingLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, alicloudRedisInstance.Status.Id),
					HavingLabels(authSecretLabels),
					HavingAnnotations(authSecretAnnotations),
				).Should(Succeed())

			Expect(authSecret.Data).To(HaveKey("primaryEndpoint"))
			Expect(authSecret.Data).To(HaveKey("host"))
			Expect(authSecret.Data).To(HaveKey("port"))
			Expect(authSecret.Data).To(HaveKey("authString"))
			Expect(authSecret.Data).To(HaveKeyWithValue("endpoint", []byte(kcpPrimaryEndpoint)))
		})

		// DELETE

		By("When AlicloudRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed())
		})
	})
})
