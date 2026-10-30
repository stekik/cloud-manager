package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Feature: SKR AlicloudVpcPeering", func() {

	It("Scenario: SKR AlicloudVpcPeering is created then deleted", func() {
		alicloudVpcPeeringName := "3a481746-6367-57d7-bbf2-f50d540eaa7d"
		alicloudVpcPeering := &cloudresourcesv1beta1.AlicloudVpcPeering{}

		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: alicloudVpcPeeringName}))

		const (
			remoteAccountId = "123456789012345678"
			remoteRegion    = "eu-central-1"
			remoteVpcId     = "vpc-abcdef1234567890"
		)

		By("When AlicloudVpcPeering is created", func() {
			Eventually(CreateAlicloudVpcPeering).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), alicloudVpcPeering,
					WithName(alicloudVpcPeeringName),
					WithAlicloudRemoteAccountId(remoteAccountId),
					WithAlicloudRemoteRegion(remoteRegion),
					WithAlicloudRemoteVpcId(remoteVpcId),
				).Should(Succeed())
		})

		By("Then KCP VpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					alicloudVpcPeering,
					NewObjActions(),
					AssertAlicloudVpcPeeringHasId(),
				).
				Should(Succeed(), "expected AlicloudVpcPeering to get status.Id, but it didn't")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("Then KCP remote Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					NewObjActions(WithName(alicloudVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load remote Network")
		})

		By("And Then KCP remote Network is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "failed to update status on KCP remote Network")
		})

		By("And Then KCP remote Network has AlicloudNetworkReference", func() {
			Expect(remoteNetwork.Spec.Network.Reference.Alicloud.AccountId).To(Equal(remoteAccountId))
			Expect(remoteNetwork.Spec.Network.Reference.Alicloud.Region).To(Equal(remoteRegion))
			Expect(remoteNetwork.Spec.Network.Reference.Alicloud.VpcId).To(Equal(remoteVpcId))
		})

		kcpVpcPeering := &cloudcontrolv1beta1.VpcPeering{}

		By("Then KCP VpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					NewObjActions(WithName(alicloudVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load KCP VpcPeering")
		})

		By("And Then KCP VpcPeering has RemoteNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Name).To(Equal(alicloudVpcPeering.Status.Id))
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has LocalNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Name).To(Equal(common.KcpNetworkKymaCommonName(kcpVpcPeering.Spec.Scope.Name)))
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has annotations", func() {
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(alicloudVpcPeering.Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(alicloudVpcPeering.Namespace))
		})

		By("When KCP VpcPeering is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					WithState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected),
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "failed to update status on KCP VpcPeering")
		})

		By("Then SKR AlicloudVpcPeering is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					alicloudVpcPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected)).
				Should(Succeed(), "expected SKR AlicloudVpcPeering to be Ready, but it didn't")
		})

		By("When SKR AlicloudVpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudVpcPeering).
				Should(Succeed(), "failed to delete SKR AlicloudVpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering, WithName(alicloudVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

		By("And Then KCP remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(alicloudVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})
	})

})
