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

package clients

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	argocd "github.com/argoproj/argo-cd/v3/pkg/apiclient"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/apis/common"
	namespacedv1alpha1 "github.com/crossplane-contrib/provider-argocd/apis/namespaced/v1alpha1"
)

// NewClient creates new argocd Client with provided argocd Configurations/Credentials.
func NewClient(opts *argocd.ClientOptions) *argocd.Client {
	var cl argocd.Client
	var err error
	cl, err = argocd.NewClient(opts)

	if err != nil {
		panic(err)
	}
	return &cl
}

// GetConfig constructs a Config that can be used to authenticate to argocd
// API by the argocd Go client
func GetConfig(ctx context.Context,
	c client.Client,
	lt LegacyTracker,
	mt ModernTracker,
	mg resource.Managed,
) (*apiclient.ClientOptions, error) {
	var (
		cfg *common.ProviderConfigSpec
		err error
	)

	switch m := mg.(type) {
	case resource.LegacyManaged:
		cfg, err = resolveProviderConfigLegacy(ctx, c, m, lt)
	case resource.ModernManaged:
		cfg, err = resolveProviderConfigModern(ctx, c, m, mt)
	default:
		return nil, errors.New("resource is not a managed")
	}

	if err != nil {
		return nil, errors.Wrap(err, "cannot resolve provider config")
	}

	return getAPIOpts(ctx, c, *cfg)
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func getAPIOpts(ctx context.Context, c client.Client, cfg common.ProviderConfigSpec) (*argocd.ClientOptions, error) {
	insecure := ptr.Deref(cfg.Insecure, false)
	plaintext := ptr.Deref(cfg.PlainText, false)

	authToken, err := authFromCredentials(ctx, c, cfg.Credentials)
	if err != nil {
		return nil, err
	}
	grpcWeb := ptr.Deref(cfg.GRPCWeb, false)
	grpcWebRoot := ptr.Deref(cfg.GRPCWebRootPath, "")

	return &argocd.ClientOptions{
		ServerAddr:      cfg.ServerAddr,
		Insecure:        insecure,
		PlainText:       plaintext,
		AuthToken:       authToken,
		GRPCWeb:         grpcWeb,
		GRPCWebRootPath: grpcWebRoot,
	}, nil
}

func authFromCredentials(ctx context.Context, c client.Client, creds common.ProviderCredentials) (string, error) { //nolint:gocyclo
	switch s := creds.Source; s {
	case xpv1.CredentialsSourceSecret:
		csr := creds.SecretRef
		if csr == nil {
			return "", errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return "", errors.Wrap(err, "cannot get credentials secret")
		}
		return string(s.Data[csr.Key]), nil
	case xpv1.CredentialsSourceFilesystem:
		fs := creds.Fs
		if fs == nil {
			return "", errors.New("no credentials fs given")
		}
		token, err := os.ReadFile(fs.Path)
		if err != nil {
			return "", errors.Wrap(err, "cannot read credentials file")
		}
		return string(token), nil
	case common.CredentialsSourceAzureWorkloadIdentity:
		options := &azidentity.WorkloadIdentityCredentialOptions{}
		if creds.AzureWorkloadIdentityOptions != nil {
			if creds.AzureWorkloadIdentityOptions.ClientID != nil {
				options.ClientID = *creds.AzureWorkloadIdentityOptions.ClientID
			}
			if creds.AzureWorkloadIdentityOptions.TenantID != nil {
				options.TenantID = *creds.AzureWorkloadIdentityOptions.TenantID
			}
			if creds.AzureWorkloadIdentityOptions.TokenFilePath != nil {
				options.TokenFilePath = *creds.AzureWorkloadIdentityOptions.TokenFilePath
			}
		}

		azcreds, err := azidentity.NewWorkloadIdentityCredential(options)
		if err != nil {
			return "", errors.Wrap(err, "failed to create workload identity credentials")
		}
		token, err := azcreds.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: creds.Audiences,
		})
		if err != nil {
			return "", errors.Wrap(err, "cannot get token from Azure")
		}
		return token.Token, nil
	default:
		return "", errors.Errorf("credentials source %s is not currently supported", s)
	}
}

func resolveProviderConfigLegacy(ctx context.Context, client kclient.Client, mg resource.LegacyManaged, lt LegacyTracker) (*common.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errProviderConfigNotSet
	}

	pc := &clusterv1alpha1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig.Error())
	}

	if err := lt.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errFailedToTrackUsage.Error())
	}

	return legacyToModernProviderConfigSpec(pc)
}

func resolveProviderConfigModern(ctx context.Context, crClient kclient.Client, mg resource.ModernManaged, mt ModernTracker) (*common.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errProviderConfigNotSet
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1alpha1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrapf(err, "referenced provider config kind %q is invalid for %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}
	pcObj, ok := pcRuntimeObj.(resource.ProviderConfig)
	if !ok {
		return nil, errors.Errorf("referenced provider config kind %q is not a provider config type %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig.Error())
	}

	var pcSpec common.ProviderConfigSpec
	switch pc := pcObj.(type) {
	case *namespacedv1alpha1.ProviderConfig:
		enrichLocalSecretRefs(pc, mg)
		pcSpec = pc.Spec
	case *namespacedv1alpha1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		// TODO(tydanny)
		return nil, errors.New("unknown")
	}

	if err := mt.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errFailedToTrackUsage.Error())
	}
	return &pcSpec, nil
}

func legacyToModernProviderConfigSpec(pc *clusterv1alpha1.ProviderConfig) (*common.ProviderConfigSpec, error) {
	// TODO(tydanny): this is hacky and potentially lossy, generate or manually implement
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec common.ProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func enrichLocalSecretRefs(pc *namespacedv1alpha1.ProviderConfig, mg resource.Managed) {
	if pc != nil && pc.Spec.Credentials.SecretRef != nil {
		pc.Spec.Credentials.SecretRef.Namespace = mg.GetNamespace()
	}
}

// StringValue converts a *string to string
func StringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// Int64Value converts a *int64 to int64
func Int64Value(ptr *int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// BoolValue converts a *bool to bool
func BoolValue(ptr *bool) bool {
	if ptr != nil {
		return *ptr
	}
	return false
}

// LateInitializeStringPtr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeStringPtr(in *string, from string) *string {
	if in == nil && from != "" {
		return &from
	}
	return in
}

// LateInitializeInt64Ptr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeInt64Ptr(in *int64, from int64) *int64 {
	if in == nil && from != 0 {
		return &from
	}
	return in
}

// IsBoolEqualToBoolPtr compares a *bool with bool
func IsBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}

// IsInt64EqualToInt64Ptr compares a *bool with bool
func IsInt64EqualToInt64Ptr(ip *int64, i int64) bool {
	if ip != nil {
		if !cmp.Equal(*ip, i) {
			return false
		}
	}
	return true
}
