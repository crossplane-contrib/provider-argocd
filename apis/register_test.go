/*
Copyright 2021 The Crossplane Authors.

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

package apis

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clusterv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	namespacev1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/namespace/v1alpha1"
)

// TestAddToSchemeRegistersNamespaceTypes guards against a regression where
// apis/namespace's ProviderConfig types were never added to the runtime
// Scheme, which caused providerconfig.NewReconciler to panic with "no
// version ... has been registered in scheme" when the namespace-scoped
// config controller started up.
func TestAddToSchemeRegistersNamespaceTypes(t *testing.T) {
	s := runtime.NewScheme()
	if err := AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme(): %v", err)
	}

	gvks := map[string]schema.GroupVersionKind{
		"cluster ProviderConfig":   clusterv1alpha1.ProviderConfigGroupVersionKind,
		"namespace ProviderConfig": namespacev1alpha1.ProviderConfigGroupVersionKind,
	}

	for name, gvk := range gvks {
		if !s.Recognizes(gvk) {
			t.Errorf("scheme does not recognize %s GVK %s", name, gvk)
			continue
		}
		if _, err := s.New(gvk); err != nil {
			t.Errorf("scheme.New(%s) for %s: %v", gvk, name, err)
		}
	}
}
