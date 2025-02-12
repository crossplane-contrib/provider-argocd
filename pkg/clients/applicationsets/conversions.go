// Package applicationsets contains APIs for ArgoCD applicationsets
package applicationsets

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane-contrib/provider-argocd/apis/applicationsets/v1alpha1"
)

// Converter helps to convert ArgoCD types to api types of this provider and vise-versa
// From & To shall both be defined for each type conversion, to prevent diverge from ArgoCD Types
// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend ExtV1JSONToRuntimeRawExtension
// goverter:struct:comment // +k8s:deepcopy-gen=false
// goverter:output:file ./zz_generated.conversion.go
// goverter:output:package github.com/crossplane-contrib/provider-argocd/pkg/clients/applicationsets
// +k8s:deepcopy-gen=false
type Converter interface {

	// goverter:ignore ServerRef
	// goverter:ignore ServerSelector
	// goverter:ignore NameRef
	// goverter:ignore NameSelector
	FromArgoDestination(in argocdv1alpha1.ApplicationDestination) v1alpha1.ApplicationDestination

	ToArgoDestination(in v1alpha1.ApplicationDestination) argocdv1alpha1.ApplicationDestination

	ToArgoApplicationSetSpec(in *v1alpha1.ApplicationSetParameters) *argocdv1alpha1.ApplicationSetSpec
	FromArgoApplicationSetSpec(in *argocdv1alpha1.ApplicationSetSpec) *v1alpha1.ApplicationSetParameters

	FromArgoApplicationSetStatus(in *argocdv1alpha1.ApplicationSetStatus) *v1alpha1.ArgoApplicationSetStatus
	ToArgoApplicationSetStatus(in *v1alpha1.ArgoApplicationSetStatus) *argocdv1alpha1.ApplicationSetStatus
}

// ExtV1JSONToRuntimeRawExtension converts an extv1.JSON into a
// *runtime.RawExtension.
func ExtV1JSONToRuntimeRawExtension(in extv1.JSON) *runtime.RawExtension {
	return &runtime.RawExtension{
		Raw: in.Raw,
	}
}
