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

	clusterSetup := setup.ClusterSetup{
		Name:             "argocd",
		Images:           images.ProviderImages{},
		ControllerConfig: v1alpha1.ControllerConfig{},
		SecretData:       nil,
		AddToSchemaFuncs: nil,
	}
	clusterSetup.Run(testenv)

	os.Exit(testenv.Run(m))
}
