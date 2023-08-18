package e2e

import (
	"context"
	"fmt"
	"github.com/maximilianbraun/xp-testing/pkg/xpenvfuncs"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"strings"
	"time"
)

const (
	argocdNamespace = "argocd"
)

func installArgocd(argocdVersion string) env.Func {
	return xpenvfuncs.Compose(
		xpenvfuncs.IgnoreMatchedErr(envfuncs.CreateNamespace(argocdNamespace), errors.IsAlreadyExists),
		installArgoManifests(argocdVersion),
		waitForArgocdToBeAvailable(argocdNamespace),
		addUserToArgocd(argocdNamespace),
		createProviderConfigSecret(argocdNamespace),
	)
}

func createProviderConfigSecret(namespace string) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		r, _ := resources.New(config.Client().RESTConfig())

		klog.V(4).Info("Setting up provider secrets with pod inside the cluster")
		errDecode := decoder.DecodeEachFile(
			ctx, os.DirFS("./setup"), "*",
			decoder.IgnoreErrorHandler(decoder.CreateHandler(r), errors.IsAlreadyExists),
			decoder.MutateNamespace(namespace),
		)

		if errDecode != nil {
			klog.Error("errDecode", errDecode)
		}

		return ctx, nil
	}
}

func addUserToArgocd(namespace string) env.Func {
	return xpenvfuncs.Compose(
		addUser(namespace),
		addRBAC(namespace),
	)
}

func addRBAC(namespace string) func(ctx context.Context, config *envconf.Config) (context.Context, error) {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		return patchConfigMap(ctx, config, "argocd-rbac-cm", namespace, `{"data":{"policy.csv":"g, provider-argocd, role:admin"}}`)
	}
}

func patchConfigMap(ctx context.Context, config *envconf.Config, name string, namespace string, patchData string) (context.Context, error) {
	configmap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	patch := k8s.Patch{
		PatchType: types.MergePatchType,
		Data:      []byte(patchData),
	}

	err := config.Client().Resources().Patch(ctx, &configmap, patch)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func addUser(namespace string) func(ctx context.Context, config *envconf.Config) (context.Context, error) {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		return patchConfigMap(ctx, config, "argocd-cm", namespace, `{"data":{"accounts.provider-argocd":"apiKey"}}`)
	}
}

func waitForArgocdToBeAvailable(namespace string) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		res := config.Client().Resources()
		res = res.WithNamespace(namespace)

		c := conditions.New(res)
		var deployments appsv1.DeploymentList

		err := res.List(ctx, &deployments)
		if err != nil {
			return nil, err
		}
		klog.V(4).Info("Waiting for Argocd to become available")
		for _, item := range deployments.Items {
			err := wait.For(
				c.DeploymentConditionMatch(&item, appsv1.DeploymentAvailable, v1.ConditionTrue),
				wait.WithTimeout(time.Minute), wait.WithImmediate())
			if err != nil {
				return nil, err
			}

		}
		klog.V(4).Info("Argocd has become available")
		return ctx, nil
	}
}

func installArgoManifests(argocdVersion string) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		// instead of downloading, can we maybe retrieve the version + manifest from the argocd go mod?
		// We could also download and do some matrix tests
		manifest, err := downloadManifest(argocdVersion)
		if err != nil {
			return ctx, err
		}
		r, err := resources.New(config.Client().RESTConfig())
		if err != nil {
			return ctx, err
		}
		err = decoder.DecodeEach(
			ctx,
			strings.NewReader(manifest),
			decoder.IgnoreErrorHandler(decoder.CreateHandler(r), errors.IsAlreadyExists),
			decoder.MutateNamespace(argocdNamespace),
		)

		if err != nil {
			return ctx, err
		}
		return ctx, nil
	}
}

// v2.7.4
func downloadManifest(argocdVersion string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", argocdVersion)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	d, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
