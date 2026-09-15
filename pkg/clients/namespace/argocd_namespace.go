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

package namespace

import (
	"context"

	argocd "github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterapis "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/apis/namespace/v1alpha1"
	clusterclients "github.com/crossplane-contrib/provider-argocd/pkg/clients/cluster"
)

// GetConfigV2 constructs a Config that can be used to authenticate to argocd
// API by the argocd Go client
func GetConfig(ctx context.Context, c client.Client, mg resource.ModernManaged) (*argocd.ClientOptions, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to argocd
// API by the argocd Go client
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.ModernManaged) (*argocd.ClientOptions, error) {
	ref := mg.GetProviderConfigReference()
	var spec *clusterapis.ProviderConfigSpec
	var usage resource.TypedProviderConfigUsage

	switch ref.Kind {
	case "", v1alpha1.ProviderConfigKind:
		pc := &v1alpha1.ProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: mg.GetNamespace(),
			Name:      ref.Name,
		}, pc); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced Provider")
		}
		spec = &pc.Spec
		usage = &v1alpha1.ProviderConfigUsage{}

	case v1alpha1.ClusterProviderConfigKind:
		pc := &v1alpha1.ClusterProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{Name: ref.Name}, pc); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced ClusterProviderConfig")
		}
		spec = &pc.Spec
		usage = &v1alpha1.ClusterProviderConfigUsage{}
	default:
		return nil, errors.Errorf("unsupported providerConfigRef kind %q", ref.Kind)
	}

	t := resource.NewProviderConfigUsageTracker(c, usage)
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}
	return clusterclients.GetClientOptions(ctx, c, spec)
}
