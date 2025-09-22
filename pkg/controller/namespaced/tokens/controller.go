package tokens

import (
	"context"
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
	atime "github.com/argoproj/pkg/time"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/features"
)

const (
	errNotToken          = "resource is not an ArgoCD Project Token"
	errGetProjectFailed  = "failed to get ArgoCD Project, check if project exists and permissions are correct"
	errGetRoleFailed     = "failed to get ArgoCD Project Role, verify role name and project configuration"
	errCreateTokenFailed = "failed to create ArgoCD Project Token, verify permissions and token configuration"
	errDeleteFailed      = "failed to delete ArgoCD Project Token, token may require manual cleanup"
	errKubeUpdateFailed  = "cannot update Argocd Project Token custom resource"
)

// Setup adds a controller that reconciles tokens.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(v1alpha1.TokenGroupKind)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newArgocdClientFn: projects.NewProjectServiceClient,
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

	if err := features.AddMRMetrics(mgr, o, &v1alpha1.TokenList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Token{}).
		WithOptions(o.ForControllerRuntime()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.TokenGroupVersionKind),
			opts...))
}

type connector struct {
	kube              client.Client
	newArgocdClientFn func(clientOpts *apiclient.ClientOptions) (io.Closer, project.ProjectServiceClient)
	usage             clients.ModernTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Token)
	if !ok {
		return nil, errors.New(errNotToken)
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
	client projects.ProjectServiceClient
	conn   io.Closer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Token)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotToken)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	projectQuery := project.ProjectQuery{
		Name: *cr.Spec.ForProvider.Project,
	}
	project, err := e.client.Get(ctx, &projectQuery)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetProjectFailed)
	}
	roles, _, err := project.GetRoleByName(cr.Spec.ForProvider.Role)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetRoleFailed)
	}
	var token argocdv1alpha1.JWTToken
	for _, t := range roles.JWTTokens {
		if t.ID == meta.GetExternalName(cr) {
			token = t
			break
		}
	}

	if token.IssuedAt == 0 {
		return managed.ExternalObservation{}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeToken(&cr.Spec.ForProvider, &token)

	cr.Status.AtProvider = v1alpha1.TokenObservation{
		IssuedAt:  token.IssuedAt,
		ExpiresAt: &token.ExpiresAt,
		ID:        &token.ID,
	}
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isTokenUpToDate(&cr.Spec.ForProvider, token),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func lateInitializeToken(p *v1alpha1.TokenParameters, r *argocdv1alpha1.JWTToken) {
	if p.ID == "" {
		p.ID = r.ID
	}
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Token)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotToken)
	}

	expiresIn, _ := parseDuration(cr.Spec.ForProvider.ExpiresIn)
	req := createRequest(cr, expiresIn)
	res, err := e.client.CreateToken(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateTokenFailed)
	}
	token := res.GetToken()

	var claims jwt.RegisteredClaims
	parser := jwt.Parser{}
	_, _, err = parser.ParseUnverified(token, &claims)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot parse token claims")
	}
	if claims.ID == "" {
		return managed.ExternalCreation{}, errors.New("token claims ID is missing")
	}
	meta.SetExternalName(cr, claims.ID)

	return managed.ExternalCreation{}, errors.Wrap(nil, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Token)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotToken)
	}

	reqDelete := &project.ProjectTokenDeleteRequest{
		Project: *cr.Spec.ForProvider.Project,
		Role:    cr.Spec.ForProvider.Role,
		Id:      *cr.Status.AtProvider.ID,
	}
	_, err := e.client.DeleteToken(ctx, reqDelete)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDeleteFailed)
	}

	expiresIn, _ := parseDuration(cr.Spec.ForProvider.ExpiresIn)
	req := createRequest(cr, expiresIn)
	res, err := e.client.CreateToken(ctx, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTokenFailed)
	}

	err = e.upsertConnectionSecret(ctx, cr, []byte(res.GetToken()))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTokenFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Token)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotToken)
	}

	req := &project.ProjectTokenDeleteRequest{
		Project: *cr.Spec.ForProvider.Project,
		Role:    cr.Spec.ForProvider.Role,
		Id:      *cr.Status.AtProvider.ID,
	}

	_, err := e.client.DeleteToken(ctx, req)
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func createRequest(cr *v1alpha1.Token, expiresIn int64) *project.ProjectTokenCreateRequest {
	req := &project.ProjectTokenCreateRequest{
		Project:   *cr.Spec.ForProvider.Project,
		Role:      cr.Spec.ForProvider.Role,
		ExpiresIn: expiresIn,
	}

	if cr.Spec.ForProvider.ID != "" {
		req.Id = cr.Spec.ForProvider.ID
	}
	if cr.Spec.ForProvider.Description != nil && *cr.Spec.ForProvider.Description != "" {
		req.Description = *cr.Spec.ForProvider.Description
	}
	return req
}

func isTokenUpToDate(p *v1alpha1.TokenParameters, r argocdv1alpha1.JWTToken) bool { //nolint:gocyclo // checking all parameters can't be reduced
	if r.IssuedAt == 0 || p.ID != r.ID {
		return false
	}

	if p.ExpiresIn == nil || *p.ExpiresIn == "0" {
		return r.ExpiresAt == 0
	}

	now := time.Now().Unix()
	if r.ExpiresAt < now {
		return false
	}

	expiresIn, err := atime.ParseDuration(*p.ExpiresIn)
	if err != nil {
		return false
	}
	if int64(expiresIn.Seconds()) != r.ExpiresAt-r.IssuedAt {
		return false
	}

	if p.RenewAfter != nil {
		renewAfter, err := atime.ParseDuration(*p.RenewAfter)
		if err != nil {
			return false
		}
		if now-r.IssuedAt > int64(renewAfter.Seconds()) {
			return false
		}
	}

	if p.RenewBefore != nil {
		renewBefore, err := atime.ParseDuration(*p.RenewBefore)
		if err != nil {
			return false
		}
		if r.ExpiresAt-now < int64(renewBefore.Seconds()) {
			return false
		}
	}

	return true
}

func parseDuration(durationStr *string) (int64, error) {
	if durationStr == nil {
		return 0, nil
	}
	duration, err := atime.ParseDuration(*durationStr)
	if err != nil {
		return 0, err
	}
	return int64(duration.Seconds()), nil
}

func (e *external) upsertConnectionSecret(ctx context.Context, token *v1alpha1.Token, data []byte) error {
	if token.GetWriteConnectionSecretToReference() == nil {
		return nil
	}
	secret := resource.LocalConnectionSecretFor(token, v1alpha1.TokenGroupVersionKind)
	secret.Data["token"] = data
	if err := e.kube.Create(ctx, secret); err != nil {
		if kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(e.kube.Update(ctx, secret), "failed to update secret: %s", secret.Name)
		}
		return errors.Wrapf(err, "failed to create secret: %s", secret.Name)
	}
	return nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return e.conn.Close()
}
