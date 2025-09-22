package applications

import (
	"maps"
	"slices"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/applications/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
)

// IsApplicationUpToDate converts ApplicationParameters to its ArgoCD Counterpart and returns if they equal
func IsApplicationUpToDate(cr *v1alpha1.ApplicationParameters, remote *argocdv1alpha1.Application) bool {
	converter := applications.NamespacedConverterImpl{}
	cluster := converter.ToArgoApplicationSpec(cr)

	opts := []cmp.Option{
		// explicitly ignore the unexported in this type instead of adding a generic allow on all type.
		// the unexported fields should not bother here, since we don't copy them or write them
		cmpopts.IgnoreUnexported(argocdv1alpha1.ApplicationDestination{}),
	}

	// Sort finalizer slices for comparison
	slices.Sort(cr.Finalizers)
	slices.Sort(remote.Finalizers)

	return cmp.Equal(*cluster, remote.Spec, opts...) && maps.Equal(cr.Annotations, remote.Annotations) && slices.Equal(cr.Finalizers, remote.Finalizers)
}

// getApplicationCondition evaluates the application status and returns appropriate Crossplane ready state
func getApplicationCondition(status *v1alpha1.ArgoApplicationStatus) xpv1.Condition {
	if status == nil {
		return xpv1.Unavailable()
	}

	// If there's an operation in progress, check if it succeeded
	if status.OperationState != nil {
		if status.OperationState.Phase != "Succeeded" {
			return xpv1.Unavailable()
		}
	}

	healthOK := true
	if status.Health.Status != "" && status.Health.Status != "Healthy" {
		healthOK = false
	}

	if healthOK {
		return xpv1.Available()
	}

	return xpv1.Unavailable()
}
