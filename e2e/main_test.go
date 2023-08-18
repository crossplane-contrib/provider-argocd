//go:build e2e

package e2e

import (
	"os"
	"testing"

	runtime "k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane-contrib/provider-argocd/apis"

	xpv1alpha1 "github.com/crossplane/crossplane/apis/pkg/v1alpha1"
	"sigs.k8s.io/e2e-framework/pkg/env"

	"github.com/maximilianbraun/xp-testing/pkg/images"
	"github.com/maximilianbraun/xp-testing/pkg/logging"
	"github.com/maximilianbraun/xp-testing/pkg/setup"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	var verbosity = 4
	logging.EnableVerboseLogging(&verbosity)
	testenv = env.NewParallel()

	key := "crossplane/provider-argocd"
	imgs := images.GetImagesFromEnvironmentOrPanic(key, &key)
	// Enhance interface for one- based providers
	clusterSetup := setup.ClusterSetup{
		Name:   "argocd",
		Images: imgs,
		ControllerConfig: &xpv1alpha1.ControllerConfig{
			Spec: xpv1alpha1.ControllerConfigSpec{
				Image: &imgs.Package,
				// Raise sync interval to speed up tests
				// add debug output, in case necessary for debugging in e.g. CI
				Args: []string{"--debug", "--sync=5s"},
			},
		},
		SecretData:        nil,
		AddToSchemaFuncs:  []func(s *runtime.Scheme) error{apis.AddToScheme},
		CrossplaneVersion: "1.13.2",
	}

	clusterSetup.Configure(testenv)
	testenv.Setup(installArgocd("v2.7.4"))
	os.Exit(testenv.Run(m))
}
