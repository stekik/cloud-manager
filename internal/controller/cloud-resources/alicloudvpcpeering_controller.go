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

package cloudresources

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/skr/alicloudvpcpeering"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

type AlicloudVpcPeeringReconcilerFactory struct{}

func (f *AlicloudVpcPeeringReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &AlicloudVpcPeeringReconciler{
		reconciler: alicloudvpcpeering.NewReconcilerFactory().New(args),
	}
}

// AlicloudVpcPeeringReconciler reconciles a AlicloudVpcPeering object
type AlicloudVpcPeeringReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudvpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudvpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudvpcpeerings/finalizers,verbs=update

func (r *AlicloudVpcPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAlicloudVpcPeeringReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&AlicloudVpcPeeringReconcilerFactory{}).
		For(&cloudresourcesv1beta1.AlicloudVpcPeering{}).
		Complete()
}
