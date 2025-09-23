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
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/applications/v1alpha1"
	namespacedv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/namespaced/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
	"github.com/crossplane-contrib/provider-argocd/pkg/features"
)

const (
	errNotApplication   = "managed resource is not a Argocd application custom resource"
	errListFailed       = "cannot list Argocd application"
	errKubeUpdateFailed = "cannot update Argocd application custom resource"
	errCreateFailed     = "cannot create Argocd application"
	errUpdateFailed     = "cannot update Argocd application"
	errDeleteFailed     = "cannot delete Argocd application"
)

// Setup adds a controller that reconciles applications.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(v1alpha1.ApplicationGroupKind)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newArgocdClientFn: applications.NewApplicationServiceClient,
			usage:             resource.NewProviderConfigUsageTracker(mgr.GetClient(), &namespacedv1alpha1.ProviderConfigUsage{}),
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithTimeout(5 * time.Minute),
		managed.WithMetricRecorder(o.MetricOptions.MRMetrics),
	}

	opts = append(opts, (features.Opts(o))...)

	if err := features.AddMRMetrics(mgr, o, &v1alpha1.ApplicationList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Application{}).
		WithOptions(o.ForControllerRuntime()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationGroupVersionKind),
			opts...))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) (io.Closer, applications.ServiceClient)
	usage             clients.ModernTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return nil, errors.New(errNotApplication)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, nil, c.usage, cr)
	if err != nil {
		return nil, err
	}

	conn, argocdClient := c.newArgocdClientFn(cfg)
	return &external{kube: c.kube, client: argocdClient, conn: conn}, nil
}

type external struct {
	kube   client.Client
	client applications.ServiceClient
	conn   io.Closer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}

	var name = meta.GetExternalName(cr)

	if name == "" {
		return managed.ExternalObservation{}, nil
	}

	appQuery := application.ApplicationQuery{
		Name:         &name,
		AppNamespace: cr.Spec.ForProvider.AppNamespace,
	}

	if cr.Spec.ForProvider.Project != "" {
		appQuery.Projects = []string{cr.Spec.ForProvider.Project}
	}

	// we have to use List() because Get() returns permission error
	var apps *argocdv1alpha1.ApplicationList
	apps, err := e.client.List(ctx, &appQuery)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errListFailed)
	}

	if len(apps.Items) == 0 {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	if len(apps.Items) > 1 {
		return managed.ExternalObservation{}, errors.New("multiple applications found")
	}

	app := &apps.Items[0]

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, app)

	cr.Status.AtProvider = generateApplicationObservation(app)
	cr.Status.SetConditions(getApplicationCondition(&cr.Status.AtProvider))

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        IsApplicationUpToDate(&cr.Spec.ForProvider, app),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}

	createRequest := generateCreateApplicationRequest(cr)

	_, err := e.client.Create(ctx, createRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApplication)
	}
	updateRequest := generateUpdateApplicationRequest(cr)
	_, err := e.client.Update(ctx, updateRequest)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotApplication)
	}
	query := application.ApplicationDeleteRequest{
		Name:              clients.StringToPtr(meta.GetExternalName(cr)),
		AppNamespace:      cr.Spec.ForProvider.AppNamespace,
		Cascade:           cr.Spec.ForProvider.DeleteCascade,
		PropagationPolicy: cr.Spec.ForProvider.DeletePropagationPolicy,
	}

	if cr.Spec.ForProvider.Project != "" {
		query.Project = &cr.Spec.ForProvider.Project
	}

	_, err := e.client.Delete(ctx, &query)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func lateInitialize(applicationParameters *v1alpha1.ApplicationParameters, app *argocdv1alpha1.Application) {
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

	converter := &applications.NamespacedConverterImpl{}
	status := converter.FromArgoApplicationStatus(&app.Status)
	return *status
}

func generateCreateApplicationRequest(cr *v1alpha1.Application) *application.ApplicationCreateRequest {
	converter := &applications.NamespacedConverterImpl{}

	spec := converter.ToArgoApplicationSpec(&cr.Spec.ForProvider)

	app := &argocdv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:        meta.GetExternalName(cr),
			Namespace:   ptr.Deref(cr.Spec.ForProvider.AppNamespace, ""),
			Annotations: cr.Spec.ForProvider.Annotations,
			Finalizers:  cr.Spec.ForProvider.Finalizers,
		},
		Spec: *spec,
	}

	repoCreateRequest := &application.ApplicationCreateRequest{
		Application: app,
	}

	return repoCreateRequest
}

func generateUpdateApplicationRequest(cr *v1alpha1.Application) *application.ApplicationUpdateRequest {
	converter := applications.NamespacedConverterImpl{}

	spec := converter.ToArgoApplicationSpec(&cr.Spec.ForProvider)

	app := &argocdv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:        meta.GetExternalName(cr),
			Namespace:   ptr.Deref(cr.Spec.ForProvider.AppNamespace, ""),
			Annotations: cr.Spec.ForProvider.Annotations,
			Finalizers:  cr.Spec.ForProvider.Finalizers,
		},
		Spec: *spec,
	}

	o := &application.ApplicationUpdateRequest{
		Application: app,
	}
	return o
}

func (e *external) Disconnect(ctx context.Context) error {
	return e.conn.Close()
}
