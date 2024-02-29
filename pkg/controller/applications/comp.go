package applications

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/applications/v1alpha1"
)

// IsApplicationUpToDate converts ApplicationParameters to its ArgoCD Counterpart and returns if they equal
func IsApplicationUpToDate(cr *v1alpha1.ApplicationParameters, remote *argocdv1alpha1.Application) bool { // nolint:gocyclo
	converter := v1alpha1.ConverterImpl{}
	cluster := converter.ToArgoApplicationSpec(cr)

	opts := []cmp.Option{
		// explicitly ignore the unexported in this type instead of adding a generic allow on all type.
		// the unexported fields should not bother here, since we don't copy them or write them
		cmpopts.IgnoreUnexported(argocdv1alpha1.ApplicationDestination{}),
	}
	return cmp.Equal(*cluster, remote.Spec, opts...) && cmp.Equal(cr.Annotations, remote.Annotations, opts...)
}
