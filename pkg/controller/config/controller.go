package config

import (
	"context"
	"fmt"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	argocd "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	accountpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/crossplane-contrib/provider-argocd/apis/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	shortWait = 30 * time.Second
	timeout   = 2 * time.Minute

	errGetPC = "cannot get ProviderConfig"

	SecretLabelKey = "argocd.crossplane.io/session-token"
)

type Reconciler struct {
	client            client.Client
	runtimeReconciler *providerconfig.Reconciler
	scheme            *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pc := v1alpha1.ProviderConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, &pc); err != nil {
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPC)
	}

	if pc.Spec.Credentials.Source == xpv1.CredentialsSourceSecret && pc.Spec.Credentials.SecretRef.Key == "username" {
		insecure := ptr.Deref(pc.Spec.Insecure, false)
		plaintext := ptr.Deref(pc.Spec.PlainText, false)

		grpcWeb := ptr.Deref(pc.Spec.GRPCWeb, false)
		grpcWebRoot := ptr.Deref(pc.Spec.GRPCWebRootPath, "")

		clientOpts := argocd.ClientOptions{
			ServerAddr:      pc.Spec.ServerAddr,
			Insecure:        insecure,
			PlainText:       plaintext,
			GRPCWeb:         grpcWeb,
			GRPCWebRootPath: grpcWebRoot,
		}

		secretRef := pc.Spec.Credentials.SecretRef.SecretReference

		sessionSecret := v1.Secret{}
		r.client.Get(ctx, types.NamespacedName{Namespace: secretRef.Namespace, Name: secretRef.Name}, &sessionSecret)

		username := string(sessionSecret.Data["username"])
		_, sessionClient, _ := apiclient.NewClientOrDie(&clientOpts).NewSessionClient()
		sessionResp, _ := sessionClient.Create(ctx, &session.SessionCreateRequest{
			Username: username,
			Password: string(sessionSecret.Data["password"]),
		})

		clientOpts.AuthToken = sessionResp.Token

		_, accountClient := apiclient.NewClientOrDie(&clientOpts).NewAccountClientOrDie()
		resp, _ := accountClient.CreateToken(ctx, &accountpkg.CreateTokenRequest{
			Name:      username,
			ExpiresIn: 0, // no expiration.
			Id: fmt.Sprintf("%s-%s", pc.Namespace, pc.Name),
		})

		tokenSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pc.Name,
				Namespace: pc.Namespace,
			},
		}
		controllerutil.CreateOrUpdate(ctx, r.client, &tokenSecret, func() error {
			controllerutil.SetOwnerReference(&pc, &tokenSecret, r.scheme)
			tokenSecret.Data = map[string][]byte{
				"token": []byte(resp.Token),
			}
			tokenSecret.ObjectMeta.Labels = map[string]string{
				SecretLabelKey: "abc",
			}
			return nil
		})
		pc.Status.TokenSecretRef = xpv1.SecretReference{
			Name:      pc.Name,
			Namespace: pc.Namespace,
		}
		r.client.Status().Update(ctx, &pc)
	}

	return r.runtimeReconciler.Reconcile(ctx, req)
}
