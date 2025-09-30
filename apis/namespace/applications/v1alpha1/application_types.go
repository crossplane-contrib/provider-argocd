/*
Copyright 2021 The Crossplane Authors.

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
//go:generate go run -modfile ../../../../tools/go.mod -tags generate github.com/mistermx/copystruct/cmd/copystruct ../../../cluster/applications/v1alpha1 zz_generated.types.copied.go ApplicationParameters,ArgoApplicationStatus
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/apis/cluster|github.com/crossplane-contrib/provider-argocd/apis/namespace|g zz_generated.types.copied.go
//go:generate sed -i s|v1\.Reference|v1.NamespacedReference|g zz_generated.types.copied.go
//go:generate sed -i s|v1\.Selector|v1.NamespacedSelector|g zz_generated.types.copied.go

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A ApplicationSpec defines the desired state of an ArgoCD Application.
type ApplicationSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ApplicationParameters `json:"forProvider"`
}

// A ApplicationStatus represents the observed state of an ArgoCD Application.
type ApplicationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ArgoApplicationStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Application is a managed resource that represents an ArgoCD Application
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application items
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// Application type metadata
var (
	ApplicationKind             = reflect.TypeOf(Application{}).Name()
	ApplicationGroupKind        = schema.GroupKind{Group: Group, Kind: ApplicationKind}.String()
	ApplicationKindAPIVersion   = ApplicationKind + "." + SchemeGroupVersion.String()
	ApplicationGroupVersionKind = SchemeGroupVersion.WithKind(ApplicationKind)
)

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
