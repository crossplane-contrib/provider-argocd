/*
Copyright 2022 The Crossplane Authors.

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

package applicationsets

import (
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-argocd/apis/applicationsets/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	appsets "github.com/crossplane-contrib/provider-argocd/pkg/clients/applicationsets"
)

const (
	errNotApplicationSet = "managed resource is not a ApplicationSet custom resource"
	errGetApplicationSet = "failed to GET ApplicationSet with ArgoCD instance"
)

// SetupApplicationSet adds a controller that reconciles ApplicationSet managed resources.
func SetupApplicationSet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ApplicationSetGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ApplicationSet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationSetGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newArgocdClientFn: appsets.NewApplicationSetServiceClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) appsets.ServiceClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSet)
	if !ok {
		return nil, errors.New(errNotApplicationSet)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newArgocdClientFn(cfg)}, nil
}

type external struct {
	kube   client.Client
	client appsets.ServiceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplicationSet)
	}

	var name = meta.GetExternalName(cr)

	if name == "" {
		return managed.ExternalObservation{}, nil
	}

	query := applicationset.ApplicationSetGetQuery{
		Name: name,
	}

	var appset *argov1alpha1.ApplicationSet

	appset, err := e.client.Get(ctx, &query)

	if err != nil && appsets.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	} else if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetApplicationSet)
	}

	current := cr.Spec.ForProvider.DeepCopy()

	cr.Status.AtProvider = generateApplicationObservation(appset)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        IsApplicationSetUpToDate(&cr.Spec.ForProvider, appset),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func generateApplicationObservation(appset *argov1alpha1.ApplicationSet) v1alpha1.ArgoApplicationSetStatus {
	converter := &v1alpha1.ConverterImpl{}
	return *converter.FromArgoApplicationSetStatus(&appset.Status)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplicationSet)
	}

	req := e.generateCreateApplicationSetRequest(cr)

	_, err := e.client.Create(ctx, req)

	return managed.ExternalCreation{}, err
}

func (e *external) generateCreateApplicationSetRequest(cr *v1alpha1.ApplicationSet) *applicationset.ApplicationSetCreateRequest {
	converter := &v1alpha1.ConverterImpl{}
	targetSpec := converter.ToArgoApplicationSetSpec(&cr.Spec.ForProvider)

	req := &applicationset.ApplicationSetCreateRequest{
		Applicationset: &argov1alpha1.ApplicationSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: meta.GetExternalName(cr),
			},
			Spec: *targetSpec,
		},
	}
	return req
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApplicationSet)
	}

	req := e.generateCreateApplicationSetRequest(cr)
	req.Upsert = true

	_, err := e.client.Create(ctx, req)

	return managed.ExternalUpdate{}, err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ApplicationSet)
	if !ok {
		return errors.New(errNotApplicationSet)
	}

	_, err := e.client.Delete(ctx, &applicationset.ApplicationSetDeleteRequest{
		Name: meta.GetExternalName(cr),
	})
	if err != nil {
		return err
	}

	return nil
}
