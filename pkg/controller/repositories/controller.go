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

package repositories

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-argocd/apis/repositories/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/repositories"
)

const (
	errNotRepository    = "managed resource is not a Argocd repository custom resource"
	errGetFailed        = "cannot get Argocd repository"
	errKubeUpdateFailed = "cannot update Argocd repository custom resource"
	errCreateFailed     = "cannot create Argocd repository"
	errUpdateFailed     = "cannot update Argocd repository"
	errDeleteFailed     = "cannot delete Argocd repository"
	errGetSecretFailed  = "cannot get Kubernetes secret"
	errFmtKeyNotFound   = "key %s is not found in referenced Kubernetes secret"
)

// SetupRepository adds a controller that reconciles repositories.
func SetupRepository(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RepositoryKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Repository{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RepositoryGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newArgocdClientFn: repositories.NewRepositoryServiceClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) repository.RepositoryServiceClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return nil, errors.New(errNotRepository)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newArgocdClientFn(cfg)}, nil
}

type external struct {
	kube   client.Client
	client repositories.RepositoryServiceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepository)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	repoQuery := repository.RepoQuery{
		Repo: meta.GetExternalName(cr),
	}

	// workaround for https://github.com/argoproj/argo-cd/issues/5951
	// we have to use ListRepositories() because Get() doesn't work -^
	repository := &argocdv1alpha1.Repository{}
	repositoryList, err := e.client.ListRepositories(ctx, &repoQuery)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	if repositoryList.Items != nil {
		for _, r := range repositoryList.Items {
			if r.Repo == repoQuery.Repo {
				repository = r
				break
			}
		}
	}
	if repository.Repo == "" {
		return managed.ExternalObservation{}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeRepository(&cr.Spec.ForProvider, repository)

	cr.Status.AtProvider = generateRepositoryObservation(repository)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isRepositoryUpToDate(&cr.Spec.ForProvider, repository),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepository)
	}

	repoCreateRequest := generateCreateRepositoryOptions(&cr.Spec.ForProvider)

	if cr.Spec.ForProvider.PasswordRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.PasswordRef)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		repoCreateRequest.Repo.Password = string(payload)
	}
	if cr.Spec.ForProvider.SSHPrivateKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.SSHPrivateKeyRef)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		repoCreateRequest.Repo.SSHPrivateKey = string(payload)
	}
	if cr.Spec.ForProvider.TLSClientCertDataRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.TLSClientCertDataRef)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		repoCreateRequest.Repo.TLSClientCertData = string(payload)
	}
	if cr.Spec.ForProvider.TLSClientCertKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.TLSClientCertKeyRef)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		repoCreateRequest.Repo.TLSClientCertKey = string(payload)
	}
	if cr.Spec.ForProvider.GithubAppPrivateKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.GithubAppPrivateKeyRef)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		repoCreateRequest.Repo.GithubAppPrivateKey = string(payload)
	}

	_, err := e.client.CreateRepository(ctx, repoCreateRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Repo)

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRepository)
	}

	repoUpdateRequest := generateUpdateRepositoryOptions(&cr.Spec.ForProvider)

	if cr.Spec.ForProvider.PasswordRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.PasswordRef)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
		repoUpdateRequest.Repo.Password = string(payload)
	}
	if cr.Spec.ForProvider.SSHPrivateKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.SSHPrivateKeyRef)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
		repoUpdateRequest.Repo.SSHPrivateKey = string(payload)
	}
	if cr.Spec.ForProvider.TLSClientCertDataRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.TLSClientCertDataRef)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
		repoUpdateRequest.Repo.TLSClientCertData = string(payload)
	}
	if cr.Spec.ForProvider.TLSClientCertKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.TLSClientCertKeyRef)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
		repoUpdateRequest.Repo.TLSClientCertKey = string(payload)
	}
	if cr.Spec.ForProvider.GithubAppPrivateKeyRef != nil {
		payload, err := e.getPayload(ctx, cr.Spec.ForProvider.GithubAppPrivateKeyRef)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
		repoUpdateRequest.Repo.GithubAppPrivateKey = string(payload)
	}

	_, err := e.client.UpdateRepository(ctx, repoUpdateRequest)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return errors.New(errNotRepository)
	}
	repoQuery := repository.RepoQuery{
		Repo: meta.GetExternalName(cr),
	}

	_, err := e.client.DeleteRepository(ctx, &repoQuery)

	return errors.Wrap(err, errDeleteFailed)
}

func lateInitializeRepository(p *v1alpha1.RepositoryParameters, r *argocdv1alpha1.Repository) { // nolint:gocyclo
	if r == nil {
		return
	}

	p.Username = clients.LateInitializeStringPtr(p.Username, r.Username)

	if p.Insecure == nil {
		p.Insecure = &r.Insecure
	}

	if p.EnableLFS == nil {
		p.EnableLFS = &r.EnableLFS
	}
	p.Type = clients.LateInitializeStringPtr(p.Type, r.Type)
	p.Name = clients.LateInitializeStringPtr(p.Name, r.Name)
	if p.InheritedCreds == nil {
		p.InheritedCreds = &r.InheritedCreds
	}
	if p.EnableOCI == nil {
		p.EnableOCI = &r.EnableLFS
	}
	p.GithubAppID = clients.LateInitializeInt64Ptr(p.GithubAppID, r.GithubAppId)
	p.GithubAppInstallationID = clients.LateInitializeInt64Ptr(p.GithubAppInstallationID, r.GithubAppInstallationId)
	p.GitHubAppEnterpriseBaseURL = clients.LateInitializeStringPtr(p.GitHubAppEnterpriseBaseURL, r.GitHubAppEnterpriseBaseURL)
}

func generateRepositoryObservation(r *argocdv1alpha1.Repository) v1alpha1.RepositoryObservation {
	if r == nil {
		return v1alpha1.RepositoryObservation{}
	}
	o := v1alpha1.RepositoryObservation{
		ConnectionState: v1alpha1.ConnectionState{
			Status:     r.ConnectionState.Status,
			Message:    r.ConnectionState.Message,
			ModifiedAt: r.ConnectionState.ModifiedAt,
		},
	}
	return o
}

func generateCreateRepositoryOptions(p *v1alpha1.RepositoryParameters) *repository.RepoCreateRequest { // nolint:gocyclo
	repo := &argocdv1alpha1.Repository{
		Repo: p.Repo,
	}
	if p.Username != nil {
		repo.Username = *p.Username
	}
	if p.Insecure != nil {
		repo.Insecure = *p.Insecure
	}
	if p.EnableLFS != nil {
		repo.EnableLFS = *p.EnableLFS
	}
	if p.Type != nil {
		repo.Type = *p.Type
	}
	if p.Name != nil {
		repo.Name = *p.Name
	}
	if p.EnableOCI != nil {
		repo.EnableOCI = *p.EnableOCI
	}
	if p.InheritedCreds != nil {
		repo.InheritedCreds = *p.InheritedCreds
	}
	if p.GithubAppID != nil {
		repo.GithubAppId = *p.GithubAppID
	}
	if p.GithubAppInstallationID != nil {
		repo.GithubAppInstallationId = *p.GithubAppInstallationID
	}
	if p.GitHubAppEnterpriseBaseURL != nil {
		repo.GitHubAppEnterpriseBaseURL = *p.GitHubAppEnterpriseBaseURL
	}

	repoCreateRequest := &repository.RepoCreateRequest{
		Repo:      repo,
		Upsert:    false,
		CredsOnly: false,
	}

	return repoCreateRequest
}

func generateUpdateRepositoryOptions(p *v1alpha1.RepositoryParameters) *repository.RepoUpdateRequest {
	repo := &argocdv1alpha1.Repository{
		Repo:           p.Repo,
		Insecure:       *p.Insecure,
		EnableLFS:      *p.EnableLFS,
		EnableOCI:      *p.EnableOCI,
		InheritedCreds: *p.InheritedCreds,
	}

	if p.Username != nil {
		repo.Username = *p.Username
	}
	if p.Type != nil {
		repo.Type = *p.Type
	}
	if p.Name != nil {
		repo.Name = *p.Name
	}
	if p.GithubAppID != nil {
		repo.GithubAppId = *p.GithubAppID
	}
	if p.GithubAppInstallationID != nil {
		repo.GithubAppInstallationId = *p.GithubAppInstallationID
	}
	if p.GitHubAppEnterpriseBaseURL != nil {
		repo.GitHubAppEnterpriseBaseURL = *p.GitHubAppEnterpriseBaseURL
	}

	o := &repository.RepoUpdateRequest{
		Repo: repo,
	}
	return o
}

func isRepositoryUpToDate(p *v1alpha1.RepositoryParameters, r *argocdv1alpha1.Repository) bool { // nolint:gocyclo

	if !cmp.Equal(p.Username, clients.StringToPtr(r.Username)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.Insecure, r.Insecure) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.EnableLFS, r.EnableLFS) {
		return false
	}
	if !cmp.Equal(p.Type, clients.StringToPtr(r.Type)) {
		return false
	}
	if !cmp.Equal(p.Name, clients.StringToPtr(r.Name)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.EnableOCI, r.EnableOCI) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.InheritedCreds, r.InheritedCreds) {
		return false
	}
	if !clients.IsInt64EqualToInt64Ptr(p.GithubAppID, r.GithubAppId) {
		return false
	}
	if !clients.IsInt64EqualToInt64Ptr(p.GithubAppInstallationID, r.GithubAppInstallationId) {
		return false
	}
	if !cmp.Equal(p.GitHubAppEnterpriseBaseURL, clients.StringToPtr(r.GitHubAppEnterpriseBaseURL)) {
		return false
	}

	return true
}

// fetch kubernetes secret payload
func (e *external) getPayload(ctx context.Context, ref *v1alpha1.SecretReference) ([]byte, error) {

	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	sc := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, sc); err != nil {
		return nil, errors.Wrap(err, errGetSecretFailed)
	}
	if ref.Key != "" {
		val, ok := sc.Data[ref.Key]
		if !ok {
			return nil, errors.New(fmt.Sprintf(errFmtKeyNotFound, ref.Key))
		}
		return val, nil
	}

	return nil, nil
}
