package applicationsets

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/applicationsets/v1alpha1"
)

// Converter helps to convert ArgoCD types to api types of this provider and vise-versa
// From & To shall both be defined for each type conversion, to prevent diverge from ArgoCD Types
// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend ExtV1JSONToRuntimeRawExtension
// goverter:extend PV1alpha1HealthStatusToPV1alpha1HealthStatus
// goverter:enum:unknown @ignore
// goverter:struct:comment // +k8s:deepcopy-gen=false
// goverter:output:file ./zz_generated.conversion.go
// goverter:output:package github.com/crossplane-contrib/provider-argocd/pkg/clients/applicationsets
// +k8s:deepcopy-gen=false
type NamespacedConverter interface {

	// goverter:ignore ServerRef
	// goverter:ignore ServerSelector
	// goverter:ignore NameRef
	// goverter:ignore NameSelector
	FromArgoDestination(in argocdv1alpha1.ApplicationDestination) v1alpha1.ApplicationDestination

	ToArgoDestination(in v1alpha1.ApplicationDestination) argocdv1alpha1.ApplicationDestination

	ToArgoApplicationSetSpec(in *v1alpha1.ApplicationSetParameters) *argocdv1alpha1.ApplicationSetSpec
	// goverter:ignore AppsetNamespace
	FromArgoApplicationSetSpec(in *argocdv1alpha1.ApplicationSetSpec) *v1alpha1.ApplicationSetParameters

	FromArgoApplicationSetStatus(in *argocdv1alpha1.ApplicationSetStatus) v1alpha1.ArgoApplicationSetStatus
	ToArgoApplicationSetStatus(in *v1alpha1.ArgoApplicationSetStatus) *argocdv1alpha1.ApplicationSetStatus
}
