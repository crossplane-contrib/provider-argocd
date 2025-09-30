/*
Copyright 2022 The Crossplane Authors.

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

package v1alpha1

// Copy types from cluster-scope apis replace references with namespace types:
//go:generate go run -modfile ../../../../tools/go.mod -tags generate github.com/mistermx/copystruct/cmd/copystruct ../../../cluster/applicationsets/v1alpha1 zz_generated.types.copied.go ApplicationSetParameters,ArgoApplicationSetStatus
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/apis/cluster|github.com/crossplane-contrib/provider-argocd/apis/namespace|g zz_generated.types.copied.go
//go:generate sed -i s|commonv1\.Reference|commonv1.NamespacedReference|g zz_generated.types.copied.go
//go:generate sed -i s|commonv1\.Selector|commonv1.NamespacedSelector|g zz_generated.types.copied.go

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A ApplicationSetSpec defines the desired state of a ApplicationSet.
type ApplicationSetSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ApplicationSetParameters `json:"forProvider"`
}

// A ApplicationSetStatus represents the observed state of a ApplicationSet.
type ApplicationSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ArgoApplicationSetStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ApplicationSet is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
type ApplicationSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSetSpec   `json:"spec"`
	Status ApplicationSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationSetList contains a list of ApplicationSet
type ApplicationSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationSet `json:"items"`
}

// ApplicationSet type metadata.
var (
	ApplicationSetKind             = reflect.TypeOf(ApplicationSet{}).Name()
	ApplicationSetGroupKind        = schema.GroupKind{Group: Group, Kind: ApplicationSetKind}.String()
	ApplicationSetKindAPIVersion   = ApplicationSetKind + "." + SchemeGroupVersion.String()
	ApplicationSetGroupVersionKind = SchemeGroupVersion.WithKind(ApplicationSetKind)
)

func init() {
	SchemeBuilder.Register(&ApplicationSet{}, &ApplicationSetList{})
}
