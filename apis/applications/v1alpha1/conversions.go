package v1alpha1

import argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

// Converter helps to convert ArgoCD types to api types of this provider and vise-versa
// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// +k8s:deepcopy-gen=false
type Converter interface {

	// goverter:ignore ServerRef
	// goverter:ignore ServerSelector
	FromArgoDestination(in argocdv1alpha1.ApplicationDestination) ApplicationDestination
	FromArgoDestinationP(in *argocdv1alpha1.ApplicationDestination) *ApplicationDestination

	// goverter:ignore ServerRef
	// goverter:ignore ServerSelector
	ToArgoDestination(in ApplicationDestination) argocdv1alpha1.ApplicationDestination
	ToArgoDestinationP(in *ApplicationDestination) *argocdv1alpha1.ApplicationDestination

	ToArgoApplicationSpec(in *ApplicationParameters) *argocdv1alpha1.ApplicationSpec

	FromArgoApplicationStatus(in *argocdv1alpha1.ApplicationStatus) *ArgoApplicationStatus
}
