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

// Package controller implements the controller managers for ArgoCD Managed Resources
package controller

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-argocd/pkg/controller/applications"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/applicationsets"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/cluster"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/config"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/controller/repositories"
)

// Setup creates all argocd API controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		config.Setup,
		repositories.SetupRepository,
		projects.SetupProject,
		cluster.SetupCluster,
		applications.SetupApplication,
		applicationsets.SetupApplicationSet,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
