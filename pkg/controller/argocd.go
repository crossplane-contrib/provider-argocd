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

package controller

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	clusterapplications "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/applications"
	clusterapplicationsets "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/applicationsets"
	clustercluster "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/cluster"
	clusterconfig "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/config"
	clusterprojects "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/projects"
	clusterrepositories "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/repositories"
	clustertokens "github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster/tokens"
	namespacedapplications "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/applications"
	namespacedapplicationsets "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/applicationsets"
	namespacedcluster "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/cluster"
	namespacedconfig "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/config"
	namespacedprojects "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/projects"
	namespacedrepositories "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/repositories"
	namespacedtokens "github.com/crossplane-contrib/provider-argocd/pkg/controller/namespaced/tokens"
)

// Setup creates all argocd API controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		clusterapplications.Setup,
		clusterapplicationsets.Setup,
		clustercluster.Setup,
		clusterconfig.Setup,
		clusterprojects.Setup,
		clusterrepositories.Setup,
		clustertokens.Setup,
		namespacedapplications.Setup,
		namespacedapplicationsets.Setup,
		namespacedcluster.Setup,
		namespacedconfig.Setup,
		namespacedprojects.Setup,
		namespacedrepositories.Setup,
		namespacedtokens.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
