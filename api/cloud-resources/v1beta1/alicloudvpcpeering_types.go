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

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AlicloudRouteTableUpdateStrategy string

const (
	AlicloudRouteTableUpdateStrategyAuto      AlicloudRouteTableUpdateStrategy = "AUTO"
	AlicloudRouteTableUpdateStrategyNone      AlicloudRouteTableUpdateStrategy = "NONE"
	AlicloudRouteTableUpdateStrategyMatched   AlicloudRouteTableUpdateStrategy = "MATCHED"
	AlicloudRouteTableUpdateStrategyUnmatched AlicloudRouteTableUpdateStrategy = "UNMATCHED"
)

// AlicloudVpcPeeringSpec defines the desired state of AlicloudVpcPeering
type AlicloudVpcPeeringSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVpcId is immutable."
	RemoteVpcId string `json:"remoteVpcId"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRegion is immutable."
	RemoteRegion string `json:"remoteRegion"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteAccountId is immutable."
	RemoteAccountId string `json:"remoteAccountId"`

	DeleteRemotePeering bool `json:"deleteRemotePeering,omitempty"`

	// +kubebuilder:default:=AUTO
	// +kubebuilder:validation:Enum=AUTO;NONE;MATCHED;UNMATCHED
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRouteTableUpdateStrategy is immutable."
	RemoteRouteTableUpdateStrategy AlicloudRouteTableUpdateStrategy `json:"remoteRouteTableUpdateStrategy,omitempty"`

	// Bandwidth for cross-region peering in Mbit/s. 0 means use controller default (1024 Mbit/s).
	// Relevant only for cross-region peerings; cross-region bandwidth is billed by AliCloud.
	// +optional
	Bandwidth int32 `json:"bandwidth,omitempty"`
}

// AlicloudVpcPeeringStatus defines the observed state of AlicloudVpcPeering
type AlicloudVpcPeeringStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AlicloudVpcPeering is the Schema for the alicloudvpcpeerings API
type AlicloudVpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudVpcPeeringSpec   `json:"spec,omitempty"`
	Status AlicloudVpcPeeringStatus `json:"status,omitempty"`
}

func (in *AlicloudVpcPeering) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AlicloudVpcPeering) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AlicloudVpcPeering) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeaturePeering
}

func (in *AlicloudVpcPeering) SpecificToProviders() []string { return []string{"alicloud"} }

func (in *AlicloudVpcPeering) State() string {
	return in.Status.State
}
func (in *AlicloudVpcPeering) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// AlicloudVpcPeeringList contains a list of AlicloudVpcPeering
type AlicloudVpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudVpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudVpcPeering{}, &AlicloudVpcPeeringList{})
}
