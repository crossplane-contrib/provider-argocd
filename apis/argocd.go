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

// Package apis contains Kubernetes API for argocd API.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	clusterapplications "github.com/crossplane-contrib/provider-argocd/apis/cluster/applications/v1alpha1"
	namespacedapplications "github.com/crossplane-contrib/provider-argocd/apis/cluster/applications/v1alpha1"
	clusterapplicationsets "github.com/crossplane-contrib/provider-argocd/apis/cluster/applicationsets/v1alpha1"
	namespacedapplicationsets "github.com/crossplane-contrib/provider-argocd/apis/cluster/applicationsets/v1alpha1"
	clustercluster "github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	namespacedcluster "github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	clusterprojects "github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
	namespacedprojects "github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
	clusterrepositories "github.com/crossplane-contrib/provider-argocd/apis/cluster/repositories/v1alpha1"
	namespacedrepositories "github.com/crossplane-contrib/provider-argocd/apis/cluster/repositories/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	namespacev1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/namespaced/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		namespacev1alpha1.SchemeBuilder.AddToScheme,
		clusterapplications.SchemeBuilder.AddToScheme,
		namespacedapplications.SchemeBuilder.AddToScheme,
		clusterapplicationsets.SchemeBuilder.AddToScheme,
		namespacedapplicationsets.SchemeBuilder.AddToScheme,
		clustercluster.SchemeBuilder.AddToScheme,
		namespacedcluster.SchemeBuilder.AddToScheme,
		clusterprojects.SchemeBuilder.AddToScheme,
		namespacedprojects.SchemeBuilder.AddToScheme,
		clusterrepositories.SchemeBuilder.AddToScheme,
		namespacedrepositories.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
