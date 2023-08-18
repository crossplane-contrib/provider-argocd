//go:build e2e

package e2e

import (
	"github.com/crossplane-contrib/provider-argocd/apis/projects/v1alpha1"
	"github.com/maximilianbraun/xp-testing/pkg/resources"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
)

func Test_Project_v1alpha1(t *testing.T) {

	resource := resources.NewResourceTestConfig(&v1alpha1.Project{}, "Project")

	fB := features.New(resource.Kind)
	fB.WithLabel("kind", resource.Kind)
	fB.Setup(resource.Setup)
	fB.Assess("create", resource.AssessCreate)
	fB.Assess("delete", resource.AssessDelete)

	testenv.Test(t, fB.Feature())

}
