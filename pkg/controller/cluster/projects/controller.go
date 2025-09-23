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

package projects

import (
	"context"
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
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

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
	clusterscopev1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/features"
)

const (
	errNotProject       = "managed resource is not a Argocd Project custom resource"
	errGetFailed        = "cannot get Argocd Project"
	errKubeUpdateFailed = "cannot update Argocd Project custom resource"
	errCreateFailed     = "cannot create Argocd Project"
	errUpdateFailed     = "cannot update Argocd Project"
	errDeleteFailed     = "cannot delete Argocd Project"
)

// Setup adds a controller that reconciles projects.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(v1alpha1.ProjectGroupKind)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newArgocdClientFn: projects.NewProjectServiceClient,
			usage:             resource.NewLegacyProviderConfigUsageTracker(mgr.GetClient(), &clusterscopev1alpha1.ProviderConfigUsage{}),
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

	if err := features.AddMRMetrics(mgr, o, &v1alpha1.ProjectList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Project{}).
		WithOptions(o.ForControllerRuntime()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ProjectGroupVersionKind),
			opts...))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) (io.Closer, project.ProjectServiceClient)
	usage             clients.LegacyTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return nil, errors.New(errNotProject)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, c.usage, nil, cr)
	if err != nil {
		return nil, err
	}

	conn, argocdClient := c.newArgocdClientFn(cfg)
	return &external{kube: c.kube, client: argocdClient, conn: conn}, nil
}

type external struct {
	kube   client.Client
	client projects.ProjectServiceClient
	conn   io.Closer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	projectQuery := project.ProjectQuery{
		Name: meta.GetExternalName(cr),
	}

	project, err := e.client.Get(ctx, &projectQuery)
	if projects.IsErrorProjectNotFound(err) {
		return managed.ExternalObservation{}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeProject(&cr.Spec.ForProvider, &project.Spec)

	cr.Status.AtProvider = generateProjectObservation(project)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isProjectUpToDate(&cr.Spec.ForProvider, project),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProject)
	}

	projCreateRequest := generateCreateProjectOptions(cr)

	resp, err := e.client.Create(ctx, projCreateRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, resp.Name)

	return managed.ExternalCreation{}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProject)
	}
	projQuery := project.ProjectQuery{
		Name: meta.GetExternalName(cr),
	}

	proj, err := e.client.Get(ctx, &projQuery)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	projUpdateRequest := generateUpdateProjectOptions(cr, proj)

	_, err = e.client.Update(ctx, projUpdateRequest)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotProject)
	}
	projQuery := project.ProjectQuery{
		Name: meta.GetExternalName(cr),
	}

	_, err := e.client.Delete(ctx, &projQuery)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func lateInitializeProject(p *v1alpha1.ProjectParameters, r *argocdv1alpha1.AppProjectSpec) { //nolint:gocyclo // checking all parameters can't be reduced
	if r == nil {
		return
	}

	if p.SourceRepos == nil {
		p.SourceRepos = r.SourceRepos
	}

	if p.Destinations == nil && r.Destinations != nil {
		p.Destinations = make([]v1alpha1.ApplicationDestination, len(r.Destinations))
		for i, res := range r.Destinations {
			res := res // FIX go linter exportloopref
			p.Destinations[i] = v1alpha1.ApplicationDestination{
				Server:    &res.Server,
				Namespace: &res.Namespace,
				Name:      &res.Name,
			}
		}
	}

	if p.Description == nil {
		p.Description = &r.Description
	}

	if p.Roles == nil && r.Roles != nil {
		p.Roles = make([]v1alpha1.ProjectRole, len(r.Roles))
		for i, res := range r.Roles {
			res := res // FIX go linter exportloopref
			jwtTokens := make([]v1alpha1.JWTToken, len(res.JWTTokens))
			for j, t := range res.JWTTokens {
				t := t // FIX go linter exportloopref
				jwtTokens[j] = v1alpha1.JWTToken{
					IssuedAt:  t.IssuedAt,
					ExpiresAt: &t.ExpiresAt,
					ID:        &t.ID,
				}
			}

			p.Roles[i] = v1alpha1.ProjectRole{
				Name:        res.Name,
				Description: &res.Description,
				Policies:    res.Policies,
				JWTTokens:   jwtTokens,
				Groups:      res.Groups,
			}
		}
	}

	if p.ClusterResourceWhitelist == nil {
		p.ClusterResourceWhitelist = r.ClusterResourceWhitelist
	}

	if p.NamespaceResourceBlacklist == nil {
		p.NamespaceResourceBlacklist = r.NamespaceResourceBlacklist
	}

	if p.OrphanedResources == nil && r.OrphanedResources != nil {
		p.OrphanedResources = &v1alpha1.OrphanedResourcesMonitorSettings{
			Warn: r.OrphanedResources.Warn,
		}
		if r.OrphanedResources.Ignore != nil {
			resourceKeys := make([]v1alpha1.OrphanedResourceKey, len(r.OrphanedResources.Ignore))
			for i, res := range r.OrphanedResources.Ignore {
				res := res // FIX go linter exportloopref
				resourceKeys[i] = v1alpha1.OrphanedResourceKey{
					Group: &res.Group,
					Kind:  &res.Kind,
					Name:  &res.Name,
				}
			}
			p.OrphanedResources.Ignore = resourceKeys
		}
	}

	if p.SyncWindows == nil && r.SyncWindows != nil {
		p.SyncWindows = make([]v1alpha1.SyncWindow, len(r.SyncWindows))

		for i, res := range r.SyncWindows {
			p.SyncWindows[i] = v1alpha1.SyncWindow{
				Kind:         ptr.To(res.Kind),
				Schedule:     ptr.To(res.Schedule),
				Duration:     ptr.To(res.Duration),
				Applications: res.Applications,
				Namespaces:   res.Namespaces,
				Clusters:     res.Clusters,
				ManualSync:   ptr.To(res.ManualSync),
			}
		}
	}

	if p.NamespaceResourceWhitelist == nil {
		p.NamespaceResourceWhitelist = r.NamespaceResourceWhitelist
	}
	if p.SignatureKeys == nil && r.SignatureKeys != nil {
		p.SignatureKeys = make([]v1alpha1.SignatureKey, len(r.SignatureKeys))
		for i, res := range r.SignatureKeys {
			p.SignatureKeys[i] = v1alpha1.SignatureKey{
				KeyID: res.KeyID,
			}
		}
	}
	if p.ClusterResourceBlacklist == nil {
		p.ClusterResourceBlacklist = r.ClusterResourceBlacklist
	}
}

func generateProjectObservation(r *argocdv1alpha1.AppProject) v1alpha1.ProjectObservation {
	if r == nil {
		return v1alpha1.ProjectObservation{}
	}

	jwtTokensByRole := make(map[string]v1alpha1.JWTTokens)
	for k, v := range r.Status.JWTTokensByRole {
		jwtTokens := make([]v1alpha1.JWTToken, len(v.Items))
		for i, t := range v.Items {
			t := t // FIX go linter exportloopref
			jwtTokens[i] = v1alpha1.JWTToken{
				IssuedAt:  t.IssuedAt,
				ExpiresAt: &t.ExpiresAt,
				ID:        &t.ID,
			}
		}

		jwtTokensByRole[k] = v1alpha1.JWTTokens{
			Items: jwtTokens,
		}
	}
	o := v1alpha1.ProjectObservation{
		JWTTokensByRole: jwtTokensByRole,
	}

	return o
}

func generateCreateProjectOptions(p *v1alpha1.Project) *project.ProjectCreateRequest {
	projSpec := generateProjectSpec(&p.Spec.ForProvider)

	projectCreateRequest := &project.ProjectCreateRequest{
		Project: &argocdv1alpha1.AppProject{
			Spec:       projSpec,
			ObjectMeta: metav1.ObjectMeta{Name: p.Name, Labels: p.Spec.ForProvider.ProjectLabels},
		},
		Upsert: false,
	}

	return projectCreateRequest
}

func generateProjectSpec(p *v1alpha1.ProjectParameters) argocdv1alpha1.AppProjectSpec { //nolint:gocyclo // checking all parameters can't be reduced
	projSpec := argocdv1alpha1.AppProjectSpec{}

	if p.SourceRepos != nil {
		projSpec.SourceRepos = p.SourceRepos
	}
	if p.Destinations != nil {
		projSpec.Destinations = make([]argocdv1alpha1.ApplicationDestination, len(p.Destinations))
		for i, r := range p.Destinations {
			projSpec.Destinations[i] = argocdv1alpha1.ApplicationDestination{
				Server:    clients.StringValue(r.Server),
				Namespace: clients.StringValue(r.Namespace),
				Name:      clients.StringValue(r.Name),
			}
		}
	}
	if p.Description != nil {
		projSpec.Description = *p.Description
	}
	if p.Roles != nil {
		projSpec.Roles = make([]argocdv1alpha1.ProjectRole, len(p.Roles))
		for i, r := range p.Roles {

			jwtTokens := make([]argocdv1alpha1.JWTToken, len(r.JWTTokens))
			for j, t := range r.JWTTokens {
				jwtTokens[j] = argocdv1alpha1.JWTToken{
					IssuedAt:  t.IssuedAt,
					ExpiresAt: clients.Int64Value(t.ExpiresAt),
					ID:        clients.StringValue(t.ID),
				}
			}

			projSpec.Roles[i] = argocdv1alpha1.ProjectRole{
				Name:        r.Name,
				Description: clients.StringValue(r.Description),
				Policies:    r.Policies,
				JWTTokens:   jwtTokens,
				Groups:      r.Groups,
			}
		}
	}
	if p.ClusterResourceWhitelist != nil {
		projSpec.ClusterResourceWhitelist = p.ClusterResourceWhitelist
	}
	if p.NamespaceResourceBlacklist != nil {
		projSpec.NamespaceResourceBlacklist = p.NamespaceResourceBlacklist
	}
	if p.OrphanedResources != nil {
		resourceKeys := make([]argocdv1alpha1.OrphanedResourceKey, len(p.OrphanedResources.Ignore))
		for i, r := range p.OrphanedResources.Ignore {
			resourceKeys[i] = argocdv1alpha1.OrphanedResourceKey{
				Group: clients.StringValue(r.Group),
				Kind:  clients.StringValue(r.Kind),
				Name:  clients.StringValue(r.Name),
			}
		}
		projSpec.OrphanedResources = &argocdv1alpha1.OrphanedResourcesMonitorSettings{
			Warn:   p.OrphanedResources.Warn,
			Ignore: resourceKeys,
		}

	}
	if p.SyncWindows != nil {
		projSpec.SyncWindows = make([]*argocdv1alpha1.SyncWindow, len(p.SyncWindows))

		for i, r := range p.SyncWindows {
			projSpec.SyncWindows[i] = &argocdv1alpha1.SyncWindow{
				Kind:         clients.StringValue(r.Kind),
				Schedule:     clients.StringValue(r.Schedule),
				Duration:     clients.StringValue(r.Duration),
				Applications: r.Applications,
				Namespaces:   r.Namespaces,
				Clusters:     r.Clusters,
				ManualSync:   clients.BoolValue(r.ManualSync),
			}
		}
	}
	if p.NamespaceResourceWhitelist != nil {
		projSpec.NamespaceResourceWhitelist = p.NamespaceResourceWhitelist
	}
	if p.SignatureKeys != nil {
		projSpec.SignatureKeys = make([]argocdv1alpha1.SignatureKey, len(p.SignatureKeys))
		for i, r := range p.SignatureKeys {
			projSpec.SignatureKeys[i] = argocdv1alpha1.SignatureKey{
				KeyID: r.KeyID,
			}
		}
	}
	if p.ClusterResourceBlacklist != nil {
		projSpec.ClusterResourceBlacklist = p.ClusterResourceBlacklist
	}

	if p.SourceNamespaces != nil {
		projSpec.SourceNamespaces = p.SourceNamespaces
	}

	return projSpec
}

func generateUpdateProjectOptions(p *v1alpha1.Project, current *argocdv1alpha1.AppProject) *project.ProjectUpdateRequest {
	projSpec := generateProjectSpec(&p.Spec.ForProvider)

	o := &project.ProjectUpdateRequest{
		Project: &argocdv1alpha1.AppProject{
			TypeMeta: p.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:            current.ObjectMeta.Name,
				ResourceVersion: current.ObjectMeta.ResourceVersion,
			},
			Spec: projSpec,
		},
	}
	return o
}

func isProjectUpToDate(p *v1alpha1.ProjectParameters, r *argocdv1alpha1.AppProject) bool {
	switch {
	case !cmp.Equal(p.SourceRepos, r.Spec.SourceRepos),
		!isEqualDestinations(p.Destinations, r.Spec.Destinations),
		clients.StringValue(p.Description) != r.Spec.Description,
		!isEqualRoles(p.Roles, r.Spec.Roles),
		!cmp.Equal(p.ClusterResourceWhitelist, r.Spec.ClusterResourceWhitelist),
		!cmp.Equal(p.NamespaceResourceBlacklist, r.Spec.NamespaceResourceBlacklist),
		!isEqualOrphanedResources(p.OrphanedResources, r.Spec.OrphanedResources),
		!isEqualSyncWindows(p.SyncWindows, r.Spec.SyncWindows),
		!cmp.Equal(p.NamespaceResourceWhitelist, r.Spec.NamespaceResourceWhitelist),
		!isEqualSignatureKeys(p.SignatureKeys, r.Spec.SignatureKeys),
		!cmp.Equal(p.ClusterResourceBlacklist, r.Spec.ClusterResourceBlacklist):
		return false
	}
	return true
}

func isEqualRoles(p []v1alpha1.ProjectRole, r []argocdv1alpha1.ProjectRole) bool {
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, role := range p {
		switch {
		case role.Name != r[i].Name,
			role.Description != nil && *role.Description != r[i].Description,
			!cmp.Equal(role.Policies, r[i].Policies),
			!cmp.Equal(role.Groups, r[i].Groups),
			!isEqualJWTTokens(role.JWTTokens, r[i].JWTTokens):
			return false
		}
	}
	return true
}

func isEqualJWTTokens(p []v1alpha1.JWTToken, r []argocdv1alpha1.JWTToken) bool {
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, jwtToken := range p {
		switch {
		case jwtToken.IssuedAt != r[i].IssuedAt,
			jwtToken.ExpiresAt != nil && *jwtToken.ExpiresAt != r[i].ExpiresAt,
			jwtToken.ID != nil && *jwtToken.ID != r[i].ID:
			return false
		}
	}
	return true
}

func isEqualDestinations(p []v1alpha1.ApplicationDestination, r []argocdv1alpha1.ApplicationDestination) bool { //nolint:gocyclo // checking all parameters can't be reduced
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, destination := range p {
		switch {
		case destination.Name != nil && *destination.Name != r[i].Name,
			destination.Namespace != nil && *destination.Namespace != r[i].Namespace,
			destination.Server != nil && *destination.Server != r[i].Server:
			return false
		}
	}
	return true
}

func isEqualOrphanedResources(p *v1alpha1.OrphanedResourcesMonitorSettings, r *argocdv1alpha1.OrphanedResourcesMonitorSettings) bool {
	if p == nil && r == nil {
		return true
	}
	if (p == nil && r != nil) || (p != nil && r == nil) {
		return false
	}
	switch {
	case *p.Warn != *r.Warn,
		!isEqualOrphanedResourceKeys(p.Ignore, r.Ignore):
		return false
	}
	return true
}

func isEqualOrphanedResourceKeys(p []v1alpha1.OrphanedResourceKey, r []argocdv1alpha1.OrphanedResourceKey) bool { //nolint:gocyclo // checking all parameters can't be reduced
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, orphanedResourceKey := range p {
		switch {
		case orphanedResourceKey.Group != nil && *orphanedResourceKey.Group != r[i].Group,
			orphanedResourceKey.Kind != nil && *orphanedResourceKey.Kind != r[i].Kind,
			orphanedResourceKey.Name != nil && *orphanedResourceKey.Name != r[i].Name:
			return false
		}
	}
	return true
}

func isEqualSignatureKeys(p []v1alpha1.SignatureKey, r []argocdv1alpha1.SignatureKey) bool {
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, signatureKey := range p {
		if signatureKey.KeyID != r[i].KeyID {
			return false
		}
	}
	return true
}

func isEqualSyncWindows(p v1alpha1.SyncWindows, r argocdv1alpha1.SyncWindows) bool { //nolint:gocyclo // checking all parameters can't be reduced
	if len(p) == 0 && r == nil {
		return true
	}
	if p == nil || r == nil || len(p) != len(r) {
		return false
	}
	for i, syncWindow := range p {
		switch {
		case syncWindow.Kind != nil && *syncWindow.Kind != r[i].Kind,
			syncWindow.Schedule != nil && *syncWindow.Schedule != r[i].Schedule,
			syncWindow.Duration != nil && *syncWindow.Duration != r[i].Duration,
			syncWindow.Applications != nil && !cmp.Equal(syncWindow.Applications, r[i].Applications),
			syncWindow.Namespaces != nil && !cmp.Equal(syncWindow.Namespaces, r[i].Namespaces),
			syncWindow.Clusters != nil && !cmp.Equal(syncWindow.Clusters, r[i].Clusters),
			syncWindow.ManualSync != nil && *syncWindow.ManualSync != r[i].ManualSync:
			return false
		}
	}
	return true
}

func (e *external) Disconnect(ctx context.Context) error {
	return e.conn.Close()
}
