//go:build e2e

package e2e

import (
	"fmt"
	"path"
	"testing"

	"github.com/crossplane-contrib/xp-testing/pkg/resources"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/crossplane-contrib/provider-argocd/apis/repositories/v1alpha1"
)

func Test_Repositories_v1alpha1(t *testing.T) {

	resource := resources.ResourceTestConfig{
		Kind:              "Repository",
		Obj:               &v1alpha1.Repository{},
		ResourceDirectory: path.Join(resources.DefaultCRFolder("Repository")),
	}

	fB := features.New(fmt.Sprintf("%v", resource.Kind))
	fB.WithLabel("kind", resource.Kind)
	fB.Setup(resource.Setup)
	fB.Assess("create", resource.AssessCreate)
	fB.Assess("delete", resource.AssessDelete)

	testenv.Test(t, fB.Feature())

}
