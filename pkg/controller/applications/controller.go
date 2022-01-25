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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-argocd/apis/applications/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
)

const (
	errNotApplication   = "managed resource is not a Argocd application custom resource"
	errGetFailed        = "cannot get Argocd application"
	errKubeUpdateFailed = "cannot update Argocd application custom resource"
	errCreateFailed     = "cannot create Argocd application"
	errUpdateFailed     = "cannot update Argocd application"
	errDeleteFailed     = "cannot delete Argocd application"
	errGetSecretFailed  = "cannot get Kubernetes secret"
	errFmtKeyNotFound   = "key %s is not found in referenced Kubernetes secret"
)

// SetupApplication adds a controller that reconciles applications.
func SetupApplication(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ApplicationKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Application{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newArgocdClientFn: applications.NewApplicationServiceClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) application.ApplicationServiceClient
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
	client applications.ApplicationServiceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	appName := meta.GetExternalName(cr)

	applicationQuery := &application.ApplicationQuery{
		Name: &appName,
	}

	application, err := e.client.Get(ctx, applicationQuery)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(applications.IsErrorApplicationNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeApplication(&cr.Spec.ForProvider, application)

	cr.Status.AtProvider = generateApplicationObservation(application)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isApplicationUpToDate(&cr.Spec.ForProvider, application),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}

	cr.Status.SetConditions(xpv1.Creating())

	app := generateCreateApplicationOptions(&cr.Spec.ForProvider)

	app.ObjectMeta.Name = cr.GetName()
	// cascade delete
	app.ObjectMeta.Finalizers = []string{"resources-finalizer.argocd.argoproj.io"}

	appCreateRequest := &application.ApplicationCreateRequest{
		Application: *app,
	}

	_, err := e.client.Create(ctx, appCreateRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// cr, ok := mg.(*v1alpha1.Application)
	// if !ok {
	// 	return managed.ExternalUpdate{}, errors.New(errNotApplication)
	// }

	// _, err := e.client.Update(ctx, generateUpdateApplicationOptions(&cr.Spec.ForProvider))
	// if err != nil {
	// 	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	// }

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return errors.New(errNotApplication)
	}
	cr.Status.SetConditions(xpv1.Deleting())

	appName := meta.GetExternalName(cr)
	appQuery := application.ApplicationDeleteRequest{
		Name: &appName,
		// Cascade: true,
	}

	_, err := e.client.Delete(ctx, &appQuery)

	return errors.Wrap(err, errDeleteFailed)
}

func lateInitializeApplication(p *v1alpha1.ApplicationParameters, a *argocdv1alpha1.Application) { // nolint:gocyclo
	if a == nil {
		return
	}

	// p.
	// 	p.Username = clients.LateInitializeStringPtr(p.Username, r.Username)

	// if p.Insecure == nil {
	// 	p.Insecure = &r.Insecure
	// }

	// if p.EnableLFS == nil {
	// 	p.EnableLFS = &r.EnableLFS
	// }
	// p.TLSClientCertData = clients.LateInitializeStringPtr(p.TLSClientCertData, r.TLSClientCertData)
	// p.TLSClientCertKey = clients.LateInitializeStringPtr(p.TLSClientCertKey, r.TLSClientCertKey)
	// p.Type = clients.LateInitializeStringPtr(p.Type, r.Type)
	// p.Name = clients.LateInitializeStringPtr(p.Name, r.Name)
	// if p.InheritedCreds == nil {
	// 	p.InheritedCreds = &r.InheritedCreds
	// }
	// if p.EnableOCI == nil {
	// 	p.EnableOCI = &r.EnableLFS
	// }
}

func generateApplicationObservation(a *argocdv1alpha1.Application) v1alpha1.ApplicationObservation {
	if a == nil {
		return v1alpha1.ApplicationObservation{}
	}
	o := v1alpha1.ApplicationObservation{
		Summary: &v1alpha1.ApplicationSummary{
			ExternalURLs: a.Status.Summary.ExternalURLs,
			Images:       a.Status.Summary.Images,
		},
		Sync: &v1alpha1.SyncStatus{
			Status:   (*string)(&a.Status.Sync.Status),
			Revision: &a.Status.Sync.Revision,
			ComparedTo: &v1alpha1.ComparedTo{
				Source: v1alpha1.ApplicationSource{
					RepoURL:        a.Status.Sync.ComparedTo.Source.RepoURL,
					Path:           &a.Status.Sync.ComparedTo.Source.Path,
					TargetRevision: &a.Status.Sync.ComparedTo.Source.TargetRevision,
					Chart:          &a.Status.Sync.ComparedTo.Source.Chart,
				},
				Destination: v1alpha1.ApplicationDestination{
					Server:    &a.Status.Sync.ComparedTo.Destination.Server,
					Namespace: &a.Status.Sync.ComparedTo.Destination.Namespace,
					Name:      &a.Status.Sync.ComparedTo.Destination.Name,
				},
			},
		},
		Health: &v1alpha1.HealthStatus{
			Status:  (*string)(&a.Status.Health.Status),
			Message: &a.Status.Health.Message,
		},
		SourceType:   (*string)(&a.Status.SourceType),
		ReconciledAt: a.Status.ReconciledAt,
		ObservedAt:   a.Status.ObservedAt,
		// TODO:
		// History
		// Conditions
		// OperationState
	}
	if len(a.Status.Resources) > 0 {
		o.Resources = make([]*v1alpha1.ResourceStatus, 0)
		for _, v := range a.Status.Resources {
			rs := v1alpha1.ResourceStatus{
				Group:           &v.Group,
				Version:         &v.Version,
				Kind:            &v.Kind,
				Namespace:       &v.Namespace,
				Name:            &v.Name,
				Status:          (*string)(&v.Status),
				Hook:            &v.Hook,
				RequiresPruning: &v.RequiresPruning,
				// Health: &v1alpha1.HealthStatus{
				// 	// Message: &v1alpha1.Health{},
				// 	// Message: &v.Health.Message,
				// 	Message: (*string)(&v.Health.Message),
				// 	Status:  (*string)(&v.Health.Status),
				// },
			}
			if v.Health != nil {
				rs.Health = &v1alpha1.HealthStatus{
					Message: (*string)(&v.Health.Message),
					Status:  (*string)(&v.Health.Status),
				}
			}
			o.Resources = append(o.Resources, &rs)
		}
	}

	return o
}

func generateCreateApplicationOptions(p *v1alpha1.ApplicationParameters) *argocdv1alpha1.Application {
	app := &argocdv1alpha1.Application{}

	if p.Destination.Name != nil {
		app.Spec.Destination.Name = *p.Destination.Name
	}
	if p.Destination.Namespace != nil {
		app.Spec.Destination.Namespace = *p.Destination.Namespace
	}
	if p.Destination.Server != nil {
		app.Spec.Destination.Server = *p.Destination.Server
	}

	app.Spec.Source.RepoURL = p.Source.RepoURL

	if p.Source.Path != nil {
		app.Spec.Source.Path = *p.Source.Path
	} else {
		app.Spec.Source.Path = "."
	}

	if p.Source.TargetRevision != nil {
		app.Spec.Source.TargetRevision = *p.Source.TargetRevision
	}

	if p.Source.Chart != nil {
		app.Spec.Source.Chart = *p.Source.Chart
	}

	// TODO ApplicationSource
	// Implement Kustomize, Ksonnet, Directory, Plugin

	if p.Source.Helm != nil {
		helm := &argocdv1alpha1.ApplicationSourceHelm{}
		if len(p.Source.Helm.FileParameters) > 0 {
			fpArr := []argocdv1alpha1.HelmFileParameter{}
			for _, v := range p.Source.Helm.FileParameters {
				fp := argocdv1alpha1.HelmFileParameter{
					Name: *v.Name,
					Path: *v.Path,
				}
				fpArr = append(fpArr, fp)
			}
			helm.FileParameters = fpArr
		}
		if len(p.Source.Helm.Parameters) > 0 {
			hpArr := []argocdv1alpha1.HelmParameter{}
			for _, v := range p.Source.Helm.Parameters {
				hp := argocdv1alpha1.HelmParameter{
					Name:        *v.Name,
					Value:       *v.Value,
					ForceString: *v.ForceString,
				}
				hpArr = append(hpArr, hp)
			}
			helm.Parameters = hpArr
		}
		if p.Source.Helm.ReleaseName != nil {
			helm.ReleaseName = *p.Source.Helm.ReleaseName
		}
		helm.ValueFiles = p.Source.Helm.ValueFiles
		helm.ValueFiles = p.Source.Helm.ValueFiles
		if p.Source.Helm.Version != nil {
			helm.Version = *p.Source.Helm.Version
		}
	}

	if p.Project != nil {
		app.Spec.Project = *p.Project
	} else {
		app.Spec.Project = "default"
	}

	if p.SyncPolicy != nil {
		syncPolicy := &argocdv1alpha1.SyncPolicy{}
		if p.SyncPolicy.Automated != nil {
			syncPolicy.Automated = &argocdv1alpha1.SyncPolicyAutomated{}
			if p.SyncPolicy.Automated.Prune != nil {
				syncPolicy.Automated.Prune = *p.SyncPolicy.Automated.Prune
				// app.Spec.SyncPolicy.Automated.Prune = syncPolicy.Automated.Prune
			}
			if p.SyncPolicy.Automated.SelfHeal != nil {
				syncPolicy.Automated.SelfHeal = *p.SyncPolicy.Automated.SelfHeal
				// app.Spec.SyncPolicy.Automated.SelfHeal = *syncPolicy.Automated.SelfHeal
			}
			if p.SyncPolicy.Automated.AllowEmpty != nil {
				syncPolicy.Automated.AllowEmpty = *p.SyncPolicy.Automated.AllowEmpty
				// app.Spec.SyncPolicy.Automated.AllowEmpty = *syncPolicy.Automated.AllowEmpty
			}
		}
		if p.SyncPolicy.SyncOptions != nil {
			for _, v := range p.SyncPolicy.SyncOptions {
				app.Spec.SyncPolicy.SyncOptions.AddOption(*v)
			}
		}
		if p.SyncPolicy.Retry != nil {
			if p.SyncPolicy.Retry.Limit != nil {
				app.Spec.SyncPolicy.Retry.Limit = *p.SyncPolicy.Retry.Limit
			}
			if p.SyncPolicy.Retry.Backoff != nil {
				if p.SyncPolicy.Retry.Backoff.Duration != nil {
					app.Spec.SyncPolicy.Retry.Backoff.Duration = *p.SyncPolicy.Retry.Backoff.Duration
				}
				if p.SyncPolicy.Retry.Backoff.Factor != nil {
					app.Spec.SyncPolicy.Retry.Backoff.Factor = p.SyncPolicy.Retry.Backoff.Factor
				}
				if p.SyncPolicy.Retry.Backoff.MaxDuration != nil {
					app.Spec.SyncPolicy.Retry.Backoff.MaxDuration = *p.SyncPolicy.Retry.Backoff.MaxDuration
				}
			}
		}
		app.Spec.SyncPolicy = syncPolicy
	}

	return app
}

// func generateUpdateApplicationOptions(p *v1alpha1.ApplicationParameters) *application.RepoUpdateRequest {
// 	repo := &argocdv1alpha1.Application{
// 		Repo:           p.Repo,
// 		Insecure:       *p.Insecure,
// 		EnableLFS:      *p.EnableLFS,
// 		EnableOCI:      *p.EnableOCI,
// 		InheritedCreds: *p.InheritedCreds,
// 	}

// 	if p.Username != nil {
// 		repo.Username = *p.Username
// 	}
// 	if p.TLSClientCertData != nil {
// 		repo.TLSClientCertData = *p.TLSClientCertData
// 	}
// 	if p.TLSClientCertKey != nil {
// 		repo.TLSClientCertKey = *p.TLSClientCertKey
// 	}
// 	if p.Type != nil {
// 		repo.Type = *p.Type
// 	}
// 	if p.Name != nil {
// 		repo.Name = *p.Name
// 	}

// 	o := &application.RepoUpdateRequest{
// 		Repo: repo,
// 	}
// 	return o
// }

func isApplicationUpToDate(p *v1alpha1.ApplicationParameters, a *argocdv1alpha1.Application) bool {

	// if !cmp.Equal(p.Username, stringToPtr(r.Username)) {
	// 	return false
	// }
	// if !isBoolEqualToBoolPtr(p.Insecure, r.Insecure) {
	// 	return false
	// }
	// if !isBoolEqualToBoolPtr(p.EnableLFS, r.EnableLFS) {
	// 	return false
	// }
	// if !cmp.Equal(p.TLSClientCertData, stringToPtr(r.TLSClientCertData)) {
	// 	return false
	// }
	// if !cmp.Equal(p.TLSClientCertKey, stringToPtr(r.TLSClientCertKey)) {
	// 	return false
	// }
	// if !cmp.Equal(p.Type, stringToPtr(r.Type)) {
	// 	return false
	// }
	// if !cmp.Equal(p.Name, stringToPtr(r.Name)) {
	// 	return false
	// }
	// if !isBoolEqualToBoolPtr(p.EnableOCI, r.EnableOCI) {
	// 	return false
	// }
	// if !isBoolEqualToBoolPtr(p.InheritedCreds, r.InheritedCreds) {
	// 	return false
	// }

	return true
}

// stringToPtr converts string to *string
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IsBoolEqualToBoolPtr compares a *bool with bool
func isBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}
