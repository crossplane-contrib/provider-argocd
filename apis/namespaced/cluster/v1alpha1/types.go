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
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterParameters define the desired state of an ArgoCD Cluster
type ClusterParameters struct {
	// Server is the API server URL of the Kubernetes cluster. Optional if using a kubeconfig
	// +optional
	Server *string `json:"server,omitempty"`
	// Name of the cluster. If omitted, will use the server address. Optional if using a kubeconfig
	// +optional
	Name *string `json:"name"`
	// Config holds cluster information for connecting to a cluster
	Config ClusterConfig `json:"config"`
	// Holds list of namespaces which are accessible in that cluster. Cluster level resources will be ignored if namespace list is not empty.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`
	// Shard contains optional shard number. Calculated on the fly by the application controller if not specified.
	// +optional
	Shard *int64 `json:"shard,omitempty"`
	// Reference between project and cluster that allow you automatically to be added as item inside Destinations project entity
	// +optional
	Project *string `json:"project,omitempty"`
	// Labels for cluster secret metadata
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations for cluster secret metadata
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ClusterConfig holds cluster information for connecting to a cluster
type ClusterConfig struct {
	// Server requires Basic authentication
	// +optional
	Username *string `json:"username,omitempty"`
	// PasswordSecretRef contains a reference to a kubernetes secret containing the Password
	// +optional
	PasswordSecretRef *SecretReference `json:"passwordSecretRef,omitempty"`
	// BearerTokenSecretRef contains a reference to a kubernetes secret containing the BearerToken
	// +optional
	BearerTokenSecretRef *SecretReference `json:"bearerTokenSecretRef,omitempty"`
	// TLSClientConfig contains settings to enable transport layer security
	// +optional
	TLSClientConfig *TLSClientConfig `json:"tlsClientConfig,omitempty"`
	// AWSAuthConfig contains IAM authentication configuration
	// +optional
	AWSAuthConfig *AWSAuthConfig `json:"awsAuthConfig,omitempty"`
	// ExecProviderConfig contains configuration for an exec provider
	// +optional
	ExecProviderConfig *ExecProviderConfig `json:"execProviderConfig,omitempty"`
	// KubeconfigSecretRef contains a reference to a Kubernetes secret entry that
	// contains a raw kubeconfig in YAML or JSON.
	// See https://kubernetes.io/docs/reference/config-api/kubeconfig.v1/ for more
	// info about Kubeconfigs
	// +optional
	KubeconfigSecretRef *SecretReference `json:"kubeconfigSecretRef,omitempty"`
}

// SecretReference holds the reference to a Kubernetes secret
type SecretReference struct {
	// Name of the secret.
	Name string `json:"name"`

	// Namespace of the secret.
	Namespace string `json:"namespace"`

	// Key whose value will be used.
	Key string `json:"key"`
}

// KubeconfigObservation holds the status of a referenced Kubeconfig
type KubeconfigObservation struct {
	Secret SecretObservation `json:"secret,omitempty"`
}

// SecretObservation observes a secret
type SecretObservation struct {
	// ResourceVersion tracks the meta1.ResourceVersion of an Object
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// ClusterInfo holds information about cluster cache and state
type ClusterInfo struct {
	// ConnectionState contains information about the connection to the cluster
	// +optional
	ConnectionState *ConnectionState `json:"connectionState,omitempty"`
	// ServerVersion contains information about the Kubernetes version of the cluster
	// +optional
	ServerVersion *string `json:"serverVersion,omitempty"`
	// CacheInfo contains information about the cluster cache
	// +optional
	CacheInfo *ClusterCacheInfo `json:"cacheInfo,omitempty"`
	// ApplicationsCount is the number of applications managed by Argo CD on the cluster
	ApplicationsCount int64 `json:"applicationsCount"`
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Insecure specifies that the server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool `json:"insecure"`
	// ServerName is passed to the server for SNI and is used in the client to check server
	// certificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	// +optional
	ServerName *string `json:"serverName,omitempty"`
	// CertDataSecretRef references a secret holding PEM-encoded bytes (typically read from a client certificate file).
	// +optional
	CertDataSecretRef *SecretReference `json:"certDataSecretRef,omitempty"`
	// KeyDataSecretRef references a secret holding PEM-encoded bytes (typically read from a client certificate key file).
	// +optional
	KeyDataSecretRef *SecretReference `json:"keyDataSecretRef,omitempty"`
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	// +optional
	CAData []byte `json:"caData,omitempty"`
	// CADataSecretRef references a secret holding PEM-encoded bytes (typically read from a root certificates bundle).
	// +optional
	CADataSecretRef *SecretReference `json:"caDataSecretRef,omitempty"`
}

// AWSAuthConfig contains IAM authentication configuration
type AWSAuthConfig struct {
	// ClusterName contains AWS cluster name
	// +optional
	ClusterName *string `json:"clusterName,omitempty"`
	// RoleARN contains optional role ARN. If set then AWS IAM Authenticator assume a role to perform cluster operations instead of the default AWS credential provider chain.
	// +optional
	RoleARN *string `json:"roleARN,omitempty"`
}

// ExecProviderConfig contains configuration for an exec provider
type ExecProviderConfig struct {
	// Command to execute
	// +optional
	Command *string `json:"command,omitempty"`
	// Arguments to pass to the command when executing it
	// +optional
	Args []string `json:"args,omitempty"`
	// Env defines additional environment variables to expose to the process
	// +optional
	Env map[string]string `json:"env,omitempty"`
	// Preferred input version of the ExecInfo
	// +optional
	APIVersion *string `json:"apiVersion,omitempty"`
	// This text is shown to the user when the executable doesn't seem to be present
	// +optional
	InstallHint *string `json:"installHint,omitempty"`
}

// ConnectionState contains information about the connection to the cluster
type ConnectionState struct {
	// Status contains the current status indicator for the connection
	Status string `json:"status"`
	// Message contains human readable information about the connection status
	Message string `json:"message"`
	// ModifiedAt contains the timestamp when this connection status has been determined
	ModifiedAt *metav1.Time `json:"attemptedAt"`
}

// ClusterCacheInfo contains information about the cluster cache
type ClusterCacheInfo struct {
	// ResourcesCount holds number of observed Kubernetes resources
	// +optional
	ResourcesCount *int64 `json:"resourcesCount,omitempty"`
	// APIsCount holds number of observed Kubernetes API count
	// +optional
	APIsCount *int64 `json:"apisCount,omitempty"`
	// LastCacheSyncTime holds time of most recent cache synchronization
	// +optional
	LastCacheSyncTime *metav1.Time `json:"lastCacheSyncTime,omitempty"`
}

// ClusterObservation represents an argocd Cluster.
type ClusterObservation struct {
	// ClusterInfo holds information about cluster cache and state
	// +optional
	ClusterInfo ClusterInfo `json:"connectionState,omitempty"`
	// Kubeconfig tracks changes to a Kubeconfig secret
	// +optional
	Kubeconfig *KubeconfigObservation `json:"kubeconfig,omitempty"`
}

// A ClusterSpec defines the desired state of an ArgoCD Cluster.
type ClusterSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ClusterParameters `json:"forProvider"`
}

// A ClusterStatus represents the observed state of an ArgoCD Cluster.
type ClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Cluster is a managed resource that represents an ArgoCD Git Cluster
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster items
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
