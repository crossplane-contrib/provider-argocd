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
	"context"
	"fmt"
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	argocdcluster "github.com/argoproj/argo-cd/v3/pkg/apiclient/cluster"
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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	clusterscopev1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/cluster"
	"github.com/crossplane-contrib/provider-argocd/pkg/features"
)

const (
	errNotCluster      = "managed resource is not a Argocd Cluster custom resource"
	errGetFailed       = "cannot get Argocd Cluster"
	errCreateFailed    = "cannot create Argocd Cluster"
	errUpdateFailed    = "cannot update Argocd Cluster"
	errDeleteFailed    = "cannot delete Argocd Cluster"
	errGetSecretFailed = "cannot get Kubernetes secret"
	errFmtKeyNotFound  = "key %s is not found in referenced Kubernetes secret"
	errParseKubeconfig = "unable to parse kubeconfig"
)

// Setup adds a controller that reconciles cluster.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(v1alpha1.ClusterGroupKind)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newArgocdClientFn: cluster.NewClusterServiceClient,
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

	if err := features.AddMRMetrics(mgr, o, &v1alpha1.ClusterList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Cluster{}).
		WithOptions(o.ForControllerRuntime()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ClusterGroupVersionKind),
			opts...))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) (io.Closer, argocdcluster.ClusterServiceClient)
	usage             clients.LegacyTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return nil, errors.New(errNotCluster)
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
	client cluster.ServiceClient
	conn   io.Closer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCluster)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	clusterQuery := argocdcluster.ClusterQuery{
		Name:   meta.GetExternalName(cr),
		Server: ptr.Deref(cr.Spec.ForProvider.Server, ""),
	}

	observedCluster, err := e.client.Get(ctx, &clusterQuery)
	if err != nil {
		switch {
		case cluster.IsErrorClusterNotFound(err):
			// Case: Cluster not found
			return managed.ExternalObservation{}, nil

		case cluster.IsErrorPermissionDenied(err):
			if meta.WasDeleted(cr) {
				// Case: Cluster is deleted,
				// and the managed resource has a deletion timestamp
				return managed.ExternalObservation{}, nil
			}
			// Case: Cluster is deleted via argocd-ui,
			// but the managed resource still exists
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil

		default:
			// Default case: Handle other errors
			return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
		}
	}
	if meta.WasDeleted(cr) && meta.GetExternalName(cr) != observedCluster.Name {
		// ArgoCD Cluster resource ignores the name field. This detects the deletion of the default cluster resource.
		return managed.ExternalObservation{}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	lateInitializeCluster(&cr.Spec.ForProvider, observedCluster)

	kubeconfigSecretResourceVersion, err := e.getSecretResourceVersion(ctx, cr.Spec.ForProvider.Config.KubeconfigSecretRef)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	currentStatusAtProvider := cr.Status.AtProvider.DeepCopy()
	cr.Status.AtProvider = generateClusterObservation(observedCluster, kubeconfigSecretResourceVersion)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isClusterUpToDate(cr, currentStatusAtProvider, observedCluster),
		ResourceLateInitialized: !cmp.Equal(currentSpec, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCluster)
	}

	clusterCreateRequest, err := e.generateCreateClusterOptions(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	resp, err := e.client.Create(ctx, clusterCreateRequest)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, resp.Name)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCluster)
	}

	clusterUpdateRequest, err := e.generateUpdateClusterOptions(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	_, err = e.client.Update(ctx, clusterUpdateRequest)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotCluster)
	}

	clusterQuery := argocdcluster.ClusterQuery{
		Server: *cr.Spec.ForProvider.Server,
		Name:   meta.GetExternalName(cr),
	}

	_, err := e.client.Delete(ctx, &clusterQuery)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func lateInitializeCluster(p *v1alpha1.ClusterParameters, r *argocdv1alpha1.Cluster) {
	if r == nil {
		return
	}

	if p.Namespaces == nil {
		p.Namespaces = r.Namespaces
	}

	if p.Shard == nil {
		p.Shard = r.Shard
	}

	if p.Server == nil {
		p.Server = &r.Server
	}

	if p.Name == nil {
		p.Name = &r.Name
	}

}

func generateClusterObservation(r *argocdv1alpha1.Cluster, kubeconfigSecretResourceVersion string) v1alpha1.ClusterObservation {
	if r == nil {
		return v1alpha1.ClusterObservation{}
	}

	o := v1alpha1.ClusterObservation{
		ClusterInfo: v1alpha1.ClusterInfo{
			ConnectionState: (*v1alpha1.ConnectionState)(&r.Info.ConnectionState),
			ServerVersion:   &r.Info.ServerVersion,
			CacheInfo: &v1alpha1.ClusterCacheInfo{
				ResourcesCount:    &r.Info.CacheInfo.ResourcesCount,
				APIsCount:         &r.Info.CacheInfo.APIsCount,
				LastCacheSyncTime: r.Info.CacheInfo.LastCacheSyncTime,
			},
			ApplicationsCount: r.Info.ApplicationsCount,
		},
	}

	if kubeconfigSecretResourceVersion != "" {
		o.Kubeconfig = &v1alpha1.KubeconfigObservation{
			Secret: v1alpha1.SecretObservation{ResourceVersion: kubeconfigSecretResourceVersion},
		}
	}

	return o
}

func (e *external) generateCreateClusterOptions(ctx context.Context, p *v1alpha1.Cluster) (*argocdcluster.ClusterCreateRequest, error) {
	argoCluster, err := e.convertClusterTypes(ctx, &p.Spec.ForProvider)
	clusterCreateRequest := &argocdcluster.ClusterCreateRequest{
		Cluster: &argoCluster,
		Upsert:  false,
	}

	return clusterCreateRequest, err
}

func (e *external) convertClusterTypes(ctx context.Context, p *v1alpha1.ClusterParameters) (argocdv1alpha1.Cluster, error) { //nolint:gocyclo // checking all parameters can't be reduced
	argoCluster := argocdv1alpha1.Cluster{
		Config: argocdv1alpha1.ClusterConfig{},
	}

	if p.Server != nil {
		argoCluster.Server = ptr.Deref(p.Server, "")
	}

	if p.Name != nil {
		argoCluster.Name = ptr.Deref(p.Name, "")
	}

	if p.Config.Username != nil {
		argoCluster.Config.Username = *p.Config.Username
	}

	if p.Config.TLSClientConfig != nil {
		if p.Config.TLSClientConfig.ServerName != nil {
			argoCluster.Config.TLSClientConfig.ServerName = *p.Config.TLSClientConfig.ServerName
		}
		argoCluster.Config.TLSClientConfig.Insecure = p.Config.TLSClientConfig.Insecure
	}

	if p.Config.AWSAuthConfig != nil {
		argoCluster.Config.AWSAuthConfig = &argocdv1alpha1.AWSAuthConfig{}

		if p.Config.AWSAuthConfig.ClusterName != nil {
			argoCluster.Config.AWSAuthConfig.ClusterName = *p.Config.AWSAuthConfig.ClusterName
		}

		if p.Config.AWSAuthConfig.RoleARN != nil {
			argoCluster.Config.AWSAuthConfig.RoleARN = *p.Config.AWSAuthConfig.RoleARN
		}
	}

	if p.Config.ExecProviderConfig != nil {
		argoCluster.Config.ExecProviderConfig = &argocdv1alpha1.ExecProviderConfig{}

		if p.Config.ExecProviderConfig.Command != nil {
			argoCluster.Config.ExecProviderConfig.Command = *p.Config.ExecProviderConfig.Command
		}

		if p.Config.ExecProviderConfig.Args != nil {
			argoCluster.Config.ExecProviderConfig.Args = p.Config.ExecProviderConfig.Args
		}

		if p.Config.ExecProviderConfig.Env != nil {
			argoCluster.Config.ExecProviderConfig.Env = p.Config.ExecProviderConfig.Env
		}

		if p.Config.ExecProviderConfig.APIVersion != nil {
			argoCluster.Config.ExecProviderConfig.APIVersion = *p.Config.ExecProviderConfig.APIVersion
		}

		if p.Config.ExecProviderConfig.InstallHint != nil {
			argoCluster.Config.ExecProviderConfig.InstallHint = *p.Config.ExecProviderConfig.InstallHint
		}
	}

	if p.Namespaces != nil {
		argoCluster.Namespaces = p.Namespaces
	}

	if p.Shard != nil {
		argoCluster.Shard = p.Shard
	}

	if p.Project != nil {
		argoCluster.Project = *p.Project
	}

	if p.Labels != nil {
		argoCluster.Labels = p.Labels
	}

	if p.Annotations != nil {
		argoCluster.Annotations = p.Annotations
	}

	err := e.resolveReferences(ctx, p, &argoCluster)

	return argoCluster, err
}

func (e *external) generateUpdateClusterOptions(ctx context.Context, p *v1alpha1.Cluster) (*argocdcluster.ClusterUpdateRequest, error) {
	clusterSpec, err := e.convertClusterTypes(ctx, &p.Spec.ForProvider)

	o := &argocdcluster.ClusterUpdateRequest{
		Cluster: &clusterSpec,
	}
	return o, err
}

func isClusterUpToDate(cr *v1alpha1.Cluster, o *v1alpha1.ClusterObservation, r *argocdv1alpha1.Cluster) bool {
	p := cr.Spec.ForProvider
	if (p.Project != nil && !cmp.Equal(*p.Project, r.Project)) || (p.Project == nil && r.Project != "") {
		return false
	}
	switch {
	case !isEqualConfig(&p.Config, &r.Config),
		!cmp.Equal(p.Namespaces, r.Namespaces),
		!cmp.Equal(p.Shard, r.Shard),
		!cmp.Equal(p.Labels, r.Labels),
		!cmp.Equal(p.Annotations, r.Annotations),
		!cmp.Equal(cr.Status.AtProvider.Kubeconfig, o.Kubeconfig):
		return false
	}

	return true
}

func isEqualConfig(p *v1alpha1.ClusterConfig, r *argocdv1alpha1.ClusterConfig) bool {
	if p == nil && r == nil {
		return true
	}
	if p == nil || r == nil {
		return false
	}
	switch {
	case p.Username != nil && *p.Username != r.Username,
		!isEqualTLSConfig(p.TLSClientConfig, &r.TLSClientConfig),
		!isEqualAWSAuthConfig(p.AWSAuthConfig, r.AWSAuthConfig),
		!isEqualExecProviderConfig(p.ExecProviderConfig, r.ExecProviderConfig):
		return false
	}
	return true
}

func isEqualTLSConfig(p *v1alpha1.TLSClientConfig, r *argocdv1alpha1.TLSClientConfig) bool {
	if p == nil && r == nil {
		return true
	}
	if p != nil && r == nil {
		return false
	}

	if p != nil && r != nil {
		switch {
		case p.Insecure != r.Insecure,
			p.ServerName != nil && *p.ServerName != r.ServerName:
			return false
		}
	}
	return true
}

func isEqualAWSAuthConfig(p *v1alpha1.AWSAuthConfig, r *argocdv1alpha1.AWSAuthConfig) bool {
	if p == nil && r == nil {
		return true
	}

	if p == nil || r == nil {
		return false
	}
	switch {
	case p.ClusterName != nil && *p.ClusterName != r.ClusterName,
		p.RoleARN != nil && *p.RoleARN != r.RoleARN:
		return false
	}
	return true
}

func isEqualExecProviderConfig(p *v1alpha1.ExecProviderConfig, r *argocdv1alpha1.ExecProviderConfig) bool {
	if p == nil && r == nil {
		return true
	}

	if p == nil || r == nil {
		return false
	}
	switch {
	case p.Command != nil && *p.Command != r.Command,
		!cmp.Equal(p.Args, r.Args),
		p.Env != nil && !cmp.Equal(p.Env, r.Env),
		p.APIVersion != nil && *p.APIVersion != r.APIVersion,
		p.InstallHint != nil && *p.InstallHint != r.InstallHint:
		return false
	}
	return true
}

func (e *external) resolveReferences(ctx context.Context, cr *v1alpha1.ClusterParameters, r *argocdv1alpha1.Cluster) error { //nolint:gocyclo // checking all parameters can't be reduced
	if cr.Config.PasswordSecretRef != nil {
		payload, err := e.getPayload(ctx, cr.Config.PasswordSecretRef)
		if err != nil {
			return err
		}
		s := string(payload)
		r.Config.Password = s
	}

	if cr.Config.BearerTokenSecretRef != nil {
		payload, err := e.getPayload(ctx, cr.Config.BearerTokenSecretRef)
		if err != nil {
			return err
		}
		s := string(payload)
		r.Config.BearerToken = s
	}

	if cr.Config.TLSClientConfig != nil && cr.Config.TLSClientConfig.CertDataSecretRef != nil {
		payload, err := e.getPayload(ctx, cr.Config.TLSClientConfig.CertDataSecretRef)
		if err != nil {
			return err
		}
		r.Config.TLSClientConfig.CertData = payload
	}

	if cr.Config.TLSClientConfig != nil {
		if cr.Config.TLSClientConfig.KeyDataSecretRef != nil {
			payload, err := e.getPayload(ctx, cr.Config.TLSClientConfig.KeyDataSecretRef)
			if err != nil {
				return err
			}
			r.Config.TLSClientConfig.KeyData = payload
		}

		if cr.Config.TLSClientConfig.CAData != nil {
			r.Config.TLSClientConfig.CAData = cr.Config.TLSClientConfig.CAData
		} else if cr.Config.TLSClientConfig.CADataSecretRef != nil {
			payload, err := e.getPayload(ctx, cr.Config.TLSClientConfig.CADataSecretRef)
			if err != nil {
				return err
			}
			r.Config.TLSClientConfig.CAData = payload
			cr.Config.TLSClientConfig.CAData = payload

		}
	}
	if cr.Config.KubeconfigSecretRef != nil {
		err := e.extractKubeconfigFromSecretRef(ctx, cr, r)

		if err != nil {
			return err
		}
	}

	return nil
}

// fetch resource version from a SecretRef so that we can track any updates
func (e *external) getSecretResourceVersion(ctx context.Context, ref *v1alpha1.SecretReference) (string, error) {
	if ref == nil {
		return "", nil
	}
	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	sc := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, sc); err != nil {
		return "", errors.Wrap(err, errGetSecretFailed)
	}
	return sc.GetResourceVersion(), nil
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

// extractKubeconfigFromSecretRef extracts login information from a Kubeconfig Secret Reference
func (e *external) extractKubeconfigFromSecretRef(ctx context.Context, p *v1alpha1.ClusterParameters, r *argocdv1alpha1.Cluster) error {

	kubeconfig, err := e.getPayload(ctx, p.Config.KubeconfigSecretRef)
	if err != nil {
		return err
	}
	restConfig, err := newRESTConfigForKubeconfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, errParseKubeconfig)
	}

	if restConfig.Host != "" {
		if p.Name == nil {
			p.Name = &restConfig.Host
			r.Name = restConfig.Host
		}
		p.Server = &restConfig.Host
		r.Server = restConfig.Host
	}

	if restConfig.BearerToken != "" {
		r.Config.BearerToken = restConfig.BearerToken
	}

	if restConfig.Username != "" {
		p.Config.Username = &restConfig.Username
		r.Config.Username = restConfig.Username
	}

	r.Config.TLSClientConfig = argocdv1alpha1.TLSClientConfig{
		Insecure:   restConfig.TLSClientConfig.Insecure,
		CAData:     restConfig.CAData,
		CertData:   restConfig.TLSClientConfig.CertData,
		KeyData:    restConfig.TLSClientConfig.KeyData,
		ServerName: restConfig.TLSClientConfig.ServerName,
	}
	return nil
}

// newRESTConfigForKubeconfig returns a REST Config for the given KubeConfigValue.
func newRESTConfigForKubeconfig(kubeConfig []byte) (*rest.Config, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return restConfig, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return e.conn.Close()
}
