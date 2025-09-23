package applicationsets

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/applicationsets/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applicationsets"
)

// IsApplicationSetUpToDate converts ApplicationParameters to its ArgoCD Counterpart and returns if they equal
func IsApplicationSetUpToDate(cr *v1alpha1.ApplicationSetParameters, remote *argocdv1alpha1.ApplicationSet) bool {
	converter := applicationsets.NamespacedConverterImpl{}
	cluster := converter.ToArgoApplicationSetSpec(cr)

	opts := []cmp.Option{
		// explicitly ignore the unexported in this type instead of adding a generic allow on all type.
		// the unexported fields should not bother here, since we don't copy them or write them
		cmpopts.IgnoreUnexported(argocdv1alpha1.ApplicationDestination{}),
		cmpopts.EquateEmpty(),
	}
	res := cmp.Equal(*cluster, remote.Spec, opts...)
	return res
}
