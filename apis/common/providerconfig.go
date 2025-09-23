// +kubebuilder:object:generate=true
package common

import xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

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
