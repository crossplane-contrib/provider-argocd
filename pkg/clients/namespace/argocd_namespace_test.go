/*
Copyright 2026 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package namespace

import (
	"context"
	"testing"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane-contrib/provider-argocd/apis"
	clusterapis "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	namespacecluster "github.com/crossplane-contrib/provider-argocd/apis/namespace/cluster/v1alpha1"
	namespaceapis "github.com/crossplane-contrib/provider-argocd/apis/namespace/v1alpha1"
)

func TestUseProviderConfigClusterProviderConfig(t *testing.T) {
	testUseProviderConfig(t, namespaceapis.ClusterProviderConfigKind)
}

func TestUseProviderConfigProviderConfig(t *testing.T) {
	testUseProviderConfig(t, namespaceapis.ProviderConfigKind)
}

func testUseProviderConfig(t *testing.T, kind string) {
	t.Helper()
	c, managed := newProviderConfigTestClient(t, kind)

	got, err := UseProviderConfig(context.Background(), c, managed)
	if err != nil {
		t.Fatalf("UseProviderConfig(): %v", err)
	}
	if got.ServerAddr != "argocd-server.argocd.svc:443" {
		t.Fatalf("ServerAddr = %q, want %q", got.ServerAddr, "argocd-server.argocd.svc:443")
	}
	if got.AuthToken != "token" {
		t.Fatalf("AuthToken = %q, want %q", got.AuthToken, "token")
	}

	usage := &namespaceapis.ProviderConfigUsage{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: managed.Namespace,
		Name:      string(managed.UID),
	}, usage); err != nil {
		t.Fatalf("Get ProviderConfigUsage(): %v", err)
	}
	if usage.ProviderConfigReference.Kind != kind {
		t.Fatalf("ProviderConfigUsage kind = %q, want %q", usage.ProviderConfigReference.Kind, kind)
	}
}

func newProviderConfigTestClient(t *testing.T, kind string) (client.Client, *namespacecluster.Cluster) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme(): %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("corev1.AddToScheme(): %v", err)
	}

	managed := &namespacecluster.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "managed-cluster",
			Namespace: "default",
			UID:       types.UID("managed-cluster-uid"),
		},
		Spec: namespacecluster.ClusterSpec{
			ManagedResourceSpec: xpv2.ManagedResourceSpec{
				ProviderConfigReference: &xpv2.ProviderConfigReference{
					Kind: kind,
					Name: "argocd-provider",
				},
			},
		},
	}
	managed.SetGroupVersionKind(namespacecluster.ClusterGroupVersionKind)

	providerConfigSpec := clusterapis.ProviderConfigSpec{
		ServerAddr: "argocd-server.argocd.svc:443",
		Credentials: clusterapis.ProviderCredentials{
			Source: xpv2.CredentialsSourceSecret,
			CommonCredentialSelectors: xpv2.CommonCredentialSelectors{
				SecretRef: &xpv2.SecretKeySelector{
					SecretReference: xpv2.SecretReference{
						Name:      "argocd-credentials",
						Namespace: "crossplane-system",
					},
					Key: "authToken",
				},
			},
		},
	}
	var providerConfig client.Object
	switch kind {
	case namespaceapis.ProviderConfigKind:
		providerConfig = &namespaceapis.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "argocd-provider", Namespace: managed.Namespace},
			Spec:       providerConfigSpec,
		}
	case namespaceapis.ClusterProviderConfigKind:
		providerConfig = &namespaceapis.ClusterProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "argocd-provider"},
			Spec:       providerConfigSpec,
		}
	default:
		t.Fatalf("unsupported provider config kind %q", kind)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "argocd-credentials", Namespace: "crossplane-system"},
		Data:       map[string][]byte{"authToken": []byte("token")},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(managed, providerConfig, secret).Build()
	return c, managed
}
