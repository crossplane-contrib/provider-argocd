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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	CredentialsSourceAzureWorkloadIdentity xpv1.CredentialsSource = "AzureWorkloadIdentity"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// ServerAddr is the hostname or IP of the argocd instance
	ServerAddr string `json:"serverAddr"`

	// PlainText specifies whether to use http vs https. Default: false.
	// +optional
	PlainText *bool `json:"plainText,omitempty"`

	// Insecure specifies whether to disable strict tls validation. Default: false.
	// +optional
	Insecure *bool `json:"insecure,omitempty"`

	// Enables gRPC-web protocol. Useful if Argo CD server is behind proxy which does not support HTTP2.
	// +optional
	GRPCWeb *bool `json:"grpcWeb,omitempty"`

	// Enables gRPC-web protocol. Useful if Argo CD server is behind proxy which does not support HTTP2. Set web root.
	// +optional
	GRPCWebRootPath *string `json:"grpcWebRootPath,omitempty"`

	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`
}

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret;Environment;Filesystem;AzureWorkloadIdentity
	Source xpv1.CredentialsSource `json:"source"`

	xpv1.CommonCredentialSelectors `json:",inline"`

	// Audiences is the audience of the token. This is used by ArgoCD to validate the token.
	// +optional
	Audiences []string `json:"audiences,omitempty"`

	// AzureWorkloadIdentityOptions contains optional parameters for AzureWorkloadIdentity.
	// +optional
	AzureWorkloadIdentityOptions *AzureWorkloadIdentityOptions `json:"azureWorkloadIdentityOptions,omitempty"`
}

// AzureWorkloadIdentityOptions contains optional parameters for AzureWorkloadIdentity.
type AzureWorkloadIdentityOptions struct {
	// ClientID of the service principal. Defaults to the value of the environment variable AZURE_CLIENT_ID.
	// +optional
	ClientID *string `json:"clientID,omitempty"`
	// TenantID of the service principal. Defaults to the value of the environment variable AZURE_TENANT_ID.
	// +optional
	TenantID *string `json:"tenantID,omitempty"`
	// TokenFilePath is the path of a file containing a Kubernetes service account token. Defaults to the value of the
	// environment variable AZURE_FEDERATED_TOKEN_FILE.
	// +optional
	TokenFilePath *string `json:"tokenFilePath,omitempty"`
}

// A ProviderConfigStatus represents the status of a ProviderConfig.
type ProviderConfigStatus struct {
	xpv1.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures how argocd controller should connect to argocd API.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,provider,argocd}
// +kubebuilder:subresource:status
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,provider,argocd}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	xpv1.ProviderConfigUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}
