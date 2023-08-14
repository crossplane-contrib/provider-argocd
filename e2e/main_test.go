//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/crossplane/crossplane/apis/pkg/v1alpha1"
	"sigs.k8s.io/e2e-framework/pkg/env"

	"github.com/maximilianbraun/xp-testing/pkg/images"
	"github.com/maximilianbraun/xp-testing/pkg/logging"
	"github.com/maximilianbraun/xp-testing/pkg/setup"
)

func TestMain(m *testing.M) {
	var verbosity = 4
	logging.EnableVerboseLogging(&verbosity)

	testenv := env.NewParallel()
	key := "crossplane/provider-argocd"
	imgs := images.GetImagesFromEnvironmentOrPanic(key, &key)
	// Enhance interface for one- based providers
	clusterSetup := setup.ClusterSetup{
		Name:   "argocd",
		Images: imgs,
		ControllerConfig: &v1alpha1.ControllerConfig{
			Spec: v1alpha1.ControllerConfigSpec{
				Image: &imgs.Package,
			},
		},
		SecretData:       nil,
		AddToSchemaFuncs: nil,
	}
	clusterSetup.Configure(testenv)

	os.Exit(testenv.Run(m))
}

func TestFoo(t *testing.T) {
	t.Skip()
}
