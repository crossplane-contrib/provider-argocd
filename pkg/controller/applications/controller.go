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

package applications

import (
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-argocd/apis/applications/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
)

const (
	errNotApplication   = "managed resource is not a Argocd application custom resource"
	errListFailed       = "cannot list Argocd application"
	errKubeUpdateFailed = "cannot update Argocd application custom resource"
	errCreateFailed     = "cannot create Argocd application"
	errUpdateFailed     = "cannot update Argocd application"
	errDeleteFailed     = "cannot delete Argocd application"
)

// SetupApplication adds a controller that reconciles applications.
func SetupApplication(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ApplicationKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Application{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newArgocdClientFn: applications.NewApplicationServiceClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) applications.ServiceClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return nil, errors.New(errNotApplication)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newArgocdClientFn(cfg)}, nil
}

type external struct {
	kube   client.Client
	client applications.ServiceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}

	var name = meta.GetExternalName(cr)

	if name == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	appQuery := application.ApplicationQuery{
		Name: &name,
	}

	// we have to use List() because Get() returns permission error
	var apps *argocdv1alpha1.ApplicationList
	apps, err := e.client.List(ctx, &appQuery)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errListFailed)
	}
	app := &argocdv1alpha1.Application{}
	for _, item := range apps.Items {
		if item.Name == name && item.Spec.Project == cr.Spec.ForProvider.Project {
			app = item.DeepCopy()
		}
	}
	if app.Name == "" {
		return managed.ExternalObservation{}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, app)

	cr.Status.AtProvider = generateApplicationObservation(app)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        IsApplicationUpToDate(&cr.Spec.ForProvider, app),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}

	createRequest := generateCreateApplicationRequest(&cr.Spec.ForProvider, cr.Name)

	_, err := e.client.Create(ctx, createRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApplication)
	}
	name := meta.GetExternalName(cr)
	updateRequest := generateUpdateRepositoryOptions(&cr.Spec.ForProvider, name)
	_, err := e.client.Update(ctx, updateRequest)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return errors.New(errNotApplication)
	}
	query := application.ApplicationDeleteRequest{
		Name: clients.StringToPtr(meta.GetExternalName(cr)),
	}

	_, err := e.client.Delete(ctx, &query)

	return errors.Wrap(err, errDeleteFailed)
}

func lateInitialize(applicationParameters *v1alpha1.ApplicationParameters, app *argocdv1alpha1.Application) { // nolint:gocyclo
	if app == nil {
		return
	}
	if applicationParameters == nil {
		return
	}
	// To be considered in future
}

func generateApplicationObservation(app *argocdv1alpha1.Application) v1alpha1.ArgoApplicationStatus {
	if app == nil {
		return v1alpha1.ArgoApplicationStatus{}
	}

	converter := v1alpha1.ConverterImpl{}
	status := converter.FromArgoApplicationStatus(&app.Status)
	return *status
}

func generateCreateApplicationRequest(p *v1alpha1.ApplicationParameters, name string) *application.ApplicationCreateRequest { // nolint:gocyclo
	converter := v1alpha1.ConverterImpl{}
	spec := converter.ToArgoApplicationSpec(p)

	app := &argocdv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: *spec,
	}

	repoCreateRequest := &application.ApplicationCreateRequest{
		Application: app,
	}

	return repoCreateRequest
}

func generateUpdateRepositoryOptions(cr *v1alpha1.ApplicationParameters, name string) *application.ApplicationUpdateRequest {
	converter := v1alpha1.ConverterImpl{}

	spec := converter.ToArgoApplicationSpec(cr)

	app := &argocdv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: *spec,
	}

	o := &application.ApplicationUpdateRequest{
		Application: app,
	}
	return o
}
