package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAlicloudVpcPeering(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AlicloudVpcPeering, opts ...ObjAction) error {
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	return clnt.Create(ctx, obj)
}

func WithAlicloudRemoteVpcId(remoteVpcId string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AlicloudVpcPeering)
			x.Spec.RemoteVpcId = remoteVpcId
		},
	}
}

func WithAlicloudRemoteRegion(remoteRegion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AlicloudVpcPeering)
			x.Spec.RemoteRegion = remoteRegion
		},
	}
}

func WithAlicloudRemoteAccountId(remoteAccountId string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AlicloudVpcPeering)
			x.Spec.RemoteAccountId = remoteAccountId
		},
	}
}

func AssertAlicloudVpcPeeringHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AlicloudVpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not AlicloudVpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the AlicloudVpcPeering ID not set")
		}
		return nil
	}
}
