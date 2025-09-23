package v1alpha1

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// ServerName returns the Spec.ForProvider.Name of an Cluster.
func ServerName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Cluster)
		if !ok {
			return ""
		}
		if r.Spec.ForProvider.Name == nil {
			return ""
		}
		return *r.Spec.ForProvider.Name
	}
}

// ServerAddress returns the Spec.ForProvider.Server of a Cluster
func ServerAddress() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Cluster)
		if !ok {
			return ""
		}
		if r.Spec.ForProvider.Server == nil {
			return ""
		}
		return *r.Spec.ForProvider.Server
	}
}
