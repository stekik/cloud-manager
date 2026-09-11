package alicloudvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAlicloudVpcPeering()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.RemoteNetwork != nil {
		return nil, ctx
	}

	remoteNetwork := &cloudcontrolv1beta1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      obj.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: obj.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NetworkSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Network: cloudcontrolv1beta1.NetworkInfo{
				Reference: &cloudcontrolv1beta1.NetworkReference{
					Alicloud: &cloudcontrolv1beta1.AlicloudNetworkReference{
						AccountId:   obj.Spec.RemoteAccountId,
						Region:      obj.Spec.RemoteRegion,
						VpcId:       obj.Spec.RemoteVpcId,
						NetworkName: obj.Spec.RemoteVpcId,
					},
				},
			},
			Type: cloudcontrolv1beta1.NetworkTypeExternal,
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, remoteNetwork)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP remote Network", composed.StopWithRequeue, ctx)
	}

	state.RemoteNetwork = remoteNetwork
	logger.Info("KCP remote Network created")

	return nil, ctx
}
