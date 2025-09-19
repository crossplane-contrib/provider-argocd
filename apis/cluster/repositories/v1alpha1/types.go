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

// RepositoryParameters define the desired state of an ArgoCD Git Repository
type RepositoryParameters struct {
	// URL of the repo
	// +immutable
	Repo string `json:"repo"`
	// Username for authenticating at the repo server
	// +optional
	Username *string `json:"username,omitempty"`
	// Password for authenticating at the repo server
	// +optional
	PasswordRef *SecretReference `json:"passwordRef,omitempty"`
	// SSH private key data for authenticating at the repo server
	// only for Git repos
	// +optional
	// SSHPrivateKey *string `json:"sshPrivateKey,omitempty"`
	SSHPrivateKeyRef *SecretReference `json:"sshPrivateKeyRef,omitempty"`
	// Whether the repo is insecure
	// +optional
	Insecure *bool `json:"insecure,omitempty"`
	// Whether git-lfs support should be enabled for this repo
	// +optional
	EnableLFS *bool `json:"enableLfs,omitempty"`
	// TLS client cert data for authenticating at the repo server
	// +optional
	TLSClientCertDataRef *SecretReference `json:"tlsClientCertDataRef,omitempty"`
	// TLS client cert key for authenticating at the repo server
	// +optional
	TLSClientCertKeyRef *SecretReference `json:"tlsClientCertKeyRef,omitempty"`
	// type of the repo, maybe "git or "helm, "git" is assumed if empty or absent
	// +optional
	Type *string `json:"type,omitempty"`
	// Project is a reference to the project with scoped repositories
	// +optional
	// only for git repos
	Project *string `json:"project,omitempty"`
	// only for Helm repos
	// +optional
	Name *string `json:"name,omitempty"`
	// Whether credentials were inherited from a credential set
	// +optional
	InheritedCreds *bool `json:"inheritedCreds,omitempty"`
	// Whether helm-oci support should be enabled for this repo
	// +optional
	EnableOCI *bool `json:"enableOCI,omitempty"`
	// Github App Private Key PEM data
	// +optional
	GithubAppPrivateKeyRef *SecretReference `json:"githubAppPrivateKeyRef,omitempty"`
	// Github App ID of the app used to access the repo
	// +optional
	GithubAppID *int64 `json:"githubAppID,omitempty"`
	// Github App Installation ID of the installed GitHub App
	// +optional
	GithubAppInstallationID *int64 `json:"githubAppInstallationID,omitempty"`
	// Github App Enterprise base url if empty will default to https://api.github.com
	// +optional
	GitHubAppEnterpriseBaseURL *string `json:"githubAppEnterpriseBaseUrl,omitempty"`
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

// PasswordObservation holds the status of a referenced password
type PasswordObservation struct {
	Secret SecretObservation `json:"secret,omitempty"`
}

// SecretObservation observes a secret
type SecretObservation struct {
	// ResourceVersion tracks the meta1.ResourceVersion of an Object
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// RepositoryObservation represents an argocd repository.
type RepositoryObservation struct {
	// Current state of repository server connecting
	ConnectionState ConnectionState `json:"connectionState,omitempty"`

	// Password tracks changes to a Password secret
	// +optional
	Password *PasswordObservation `json:"password,omitempty"`

	// SSHPrivateKey tracks changes to a SSHPrivateKey secret
	// +optional
	SSHPrivateKey *PasswordObservation `json:"sshPrivateKey,omitempty"`

	// TLSClientCertData tracks changes to a TLSClientCertData secret
	// +optional
	TLSClientCertData *PasswordObservation `json:"tlsClientCertData,omitempty"`

	// TLSClientCertKey tracks changes to a TLSClientCertKey secret
	// +optional
	TLSClientCertKey *PasswordObservation `json:"tlsClientCertKey,omitempty"`

	// GithubAppPrivateKey tracks changes to a GithubAppPrivateKey secret
	// +optional
	GithubAppPrivateKey *PasswordObservation `json:"githubAppPrivateKey,omitempty"`
}

// ConnectionState is the observed state of the argocd repository
type ConnectionState struct {
	Status     string       `json:"status,omitempty"`
	Message    string       `json:"message,omitempty"`
	ModifiedAt *metav1.Time `json:"attemptedAt,omitempty"`
}

// A RepositorySpec defines the desired state of an ArgoCD Repository.
type RepositorySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RepositoryParameters `json:"forProvider"`
}

// A RepositoryStatus represents the observed state of an ArgoCD Repository.
type RepositoryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepositoryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Repository is a managed resource that represents an ArgoCD Git Repository
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,argocd}
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepositoryList contains a list of Repository items
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}
