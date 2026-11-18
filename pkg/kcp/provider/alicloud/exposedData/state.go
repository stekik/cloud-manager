package exposedData

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclientconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/exposedData/client"
	alicloudmetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/metrics"
	alicloudutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/util"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func NewStateFactory(clientProvider alicloudexposeddataclient.ClientProvider) StateFactory {
	return &stateFactory{clientProvider: clientProvider}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState scopetypes.State) (context.Context, composed.State, error)
}

var _ StateFactory = &stateFactory{}

type stateFactory struct {
	clientProvider alicloudexposeddataclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, baseState scopetypes.State) (context.Context, composed.State, error) {
	scope := baseState.ObjAsScope()
	accountId := alicloudmetrics.AccountIdFromScope(scope)
	region := scope.Spec.Region

	assumeRoleArn := alicloudutil.RoleArnDefault(accountId)

	ctx = alicloudmetrics.AccountIdIntoContext(ctx, accountId)

	c, err := f.clientProvider(
		ctx,
		region,
		alicloudclientconfig.AlicloudConfig.AccessKeyId,
		alicloudclientconfig.AlicloudConfig.AccessKeySecret,
		assumeRoleArn,
	)
	if err != nil {
		return ctx, nil, err
	}

	return ctx, newState(baseState, c), nil
}

func newState(baseState scopetypes.State, client alicloudexposeddataclient.Client) *State {
	return &State{
		State:  baseState,
		client: client,
	}
}

type State struct {
	scopetypes.State

	client      alicloudexposeddataclient.Client
	natGateways []alicloudexposeddataclient.NatGatewayInfo
}
