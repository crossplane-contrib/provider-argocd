//go:build e2e

package e2e

import (
	"fmt"
	"github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	"github.com/maximilianbraun/xp-testing/pkg/resources"
	"path"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
)

// Example how to use subdirectories
func Test_Cluster_v1alpha1_simple(t *testing.T) {

	resource := resources.ResourceTestConfig{
		Kind:              "Cluster",
		Obj:               &v1alpha1.Cluster{},
		ResourceDirectory: path.Join(resources.DefaultCRFolder("Cluster"), "simple"),
	}

	fB := features.New(fmt.Sprintf("%v-%v", resource.Kind, "simple"))
	fB.WithLabel("kind", resource.Kind)
	fB.Setup(resource.Setup)
	fB.Assess("create", resource.AssessCreate)
	fB.Assess("delete", resource.AssessDelete)

	testenv.Test(t, fB.Feature())

}
