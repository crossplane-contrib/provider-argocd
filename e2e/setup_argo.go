package e2e

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/crossplane-contrib/xp-testing/pkg/xpenvfuncs"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
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
		if downstrCtx, err := waitForConfigMapAvailable(ctx, config, "argocd-rbac-cm", namespace); err != nil {
			return downstrCtx, err
		}
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
		_, err := waitForConfigMapAvailable(ctx, config, "argocd-cm", namespace)
		if err != nil {
			return ctx, err

		}
		return patchConfigMap(ctx, config, "argocd-cm", namespace, `{"data":{"accounts.provider-argocd":"apiKey"}}`)
	}
}

func waitForConfigMapAvailable(ctx context.Context, config *envconf.Config, name string, namespace string) (context.Context, error) {
	res := config.Client().Resources()
	res = res.WithNamespace(namespace)

	c := conditions.New(res)
	list := v1.ConfigMapList{}
	list.Items = []v1.ConfigMap{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}}
	klog.V(4).Infof("Waiting for Configmap %s/%s", namespace, name)
	err := wait.For(c.ResourcesFound(&list), wait.WithImmediate(), wait.WithTimeout(time.Minute*2))

	return ctx, err
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
		for i := range deployments.Items {
			err := wait.For(
				c.DeploymentConditionMatch(&deployments.Items[i], appsv1.DeploymentAvailable, v1.ConditionTrue),
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

func downloadManifest(argocdVersion string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", argocdVersion)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	d, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", err
	}
	manifests := string(d)
	return manifests, nil
}
