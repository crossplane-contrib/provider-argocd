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
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	argocd "github.com/argoproj/argo-cd/v3/pkg/apiclient"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1"
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
func GetConfig(ctx context.Context, c client.Client, mg resource.LegacyManaged) (*argocd.ClientOptions, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.LegacyManaged) (*argocd.ClientOptions, error) {
	pc := &v1alpha1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewLegacyProviderConfigUsageTracker(c, &v1alpha1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	insecure := ptr.Deref(pc.Spec.Insecure, false)
	plaintext := ptr.Deref(pc.Spec.PlainText, false)

	authToken, err := authFromCredentials(ctx, c, pc.Spec.Credentials)
	if err != nil {
		return nil, err
	}
	grpcWeb := ptr.Deref(pc.Spec.GRPCWeb, false)
	grpcWebRoot := ptr.Deref(pc.Spec.GRPCWebRootPath, "")

	return &argocd.ClientOptions{
		ServerAddr:      pc.Spec.ServerAddr,
		Insecure:        insecure,
		PlainText:       plaintext,
		AuthToken:       authToken,
		GRPCWeb:         grpcWeb,
		GRPCWebRootPath: grpcWebRoot,
	}, nil
}

func authFromCredentials(ctx context.Context, c client.Client, creds v1alpha1.ProviderCredentials) (string, error) { //nolint:gocyclo
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
	case v1alpha1.CredentialsSourceAzureWorkloadIdentity:
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

// StringToPtr converts string to *string
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
