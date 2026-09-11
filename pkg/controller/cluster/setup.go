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

package cluster

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	applicationsv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/applications/v1alpha1"
	applicationsetsv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/applicationsets/v1alpha1"
	clusterv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
	repositoriesv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/repositories/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/applications"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/applicationsets"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/cluster"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/config"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/repositories"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/tokens"
)

// Setup creates all argocd API controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		config.Setup,
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, repositories.Setup, repositoriesv1alpha1.RepositoryGroupVersionKind)
		},
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, projects.Setup, projectsv1alpha1.ProjectGroupVersionKind)
		},
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, cluster.Setup, clusterv1alpha1.ClusterGroupVersionKind)
		},
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, applications.Setup, applicationsv1alpha1.ApplicationGroupVersionKind)
		},
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, applicationsets.Setup, applicationsetsv1alpha1.ApplicationSetGroupVersionKind)
		},
		func(m ctrl.Manager, o xpcontroller.Options) error {
			return setupGated(m, o, tokens.Setup, projectsv1alpha1.TokenGroupVersionKind)
		},
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

func setupGated(mgr ctrl.Manager, o xpcontroller.Options, setup func(ctrl.Manager, xpcontroller.Options) error, gvk schema.GroupVersionKind) error {
	if o.Gate == nil {
		return setup(mgr, o)
	}
	o.Gate.Register(func() {
		if err := setup(mgr, o); err != nil {
			panic(err)
		}
	}, gvk)
	return nil
}
