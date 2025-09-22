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

// ProjectParameters define the desired state of an ArgoCD Git Project
type ProjectParameters struct {
	// SourceRepos contains list of repository URLs which can be used for deployment
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-argocd/apis/namespaced/repositories/v1alpha1.Repository
	// +crossplane:generate:reference:refFieldName=SourceReposRefs
	// +crossplane:generate:reference:selectorFieldName=SourceReposSelector
	// +optional
	SourceRepos []string `json:"sourceRepos,omitempty"`
	// SourceReposRefs is a reference to an array of Repository used to set SourceRepos
	// +optional
	SourceReposRefs []xpv1.NamespacedReference `json:"sourceReposRefs,omitempty"`
	// SourceReposSelector selects references to Repositories used to set SourceRepos
	// +optional
	SourceReposSelector *xpv1.NamespacedSelector `json:"sourceReposSelector,omitempty"`
	// Destinations contains list of destinations available for deployment
	// +optional
	Destinations []ApplicationDestination `json:"destinations,omitempty"`
	// SourceNamespaces contains list of namespaces which are authorized in the project
	// +optional
	SourceNamespaces []string `json:"sourceNamespaces,omitempty"`
	// Description contains optional project description
	// +optional
	Description *string `json:"description,omitempty"`
	// Roles are user defined RBAC roles associated with this project
	// +optional
	Roles []ProjectRole `json:"roles,omitempty"`
	// ClusterResourceWhitelist contains list of whitelisted cluster level resources
	// +optional
	ClusterResourceWhitelist []metav1.GroupKind `json:"clusterResourceWhitelist,omitempty"`
	// NamespaceResourceBlacklist contains list of blacklisted namespace level resources
	// +optional
	NamespaceResourceBlacklist []metav1.GroupKind `json:"namespaceResourceBlacklist,omitempty"`
	// OrphanedResources specifies if controller should monitor orphaned resources of apps in this project
	// +optional
	OrphanedResources *OrphanedResourcesMonitorSettings `json:"orphanedResources,omitempty"`
	// SyncWindows controls when syncs can be run for apps in this project
	// +optional
	SyncWindows SyncWindows `json:"syncWindows,omitempty"`
	// NamespaceResourceWhitelist contains list of whitelisted namespace level resources
	// +optional
	NamespaceResourceWhitelist []metav1.GroupKind `json:"namespaceResourceWhitelist,omitempty"`
	// SignatureKeys contains a list of PGP key IDs that commits in Git must be signed with in order to be allowed for sync
	// +optional
	SignatureKeys []SignatureKey `json:"signatureKeys,omitempty"`
	// ClusterResourceBlacklist contains list of blacklisted cluster level resources
	// +optional
	ClusterResourceBlacklist []metav1.GroupKind `json:"clusterResourceBlacklist,omitempty"`
	// ProjectLabels labels that will be applied to the AppProject
	// +optional
	ProjectLabels map[string]string `json:"projectLabels,omitempty"`
}

// ApplicationDestination holds information about the application's destination
type ApplicationDestination struct {
	// Server specifies the URL of the target cluster and must be set to the Kubernetes control plane API
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-argocd/apis/namespaced/cluster/v1alpha1.Cluster
	// +crossplane:generate:reference:refFieldName=ServerRef
	// +crossplane:generate:reference:selectorFieldName=ServerSelector
	// +optional
	Server *string `json:"server,omitempty"`
	// ServerRef is a reference to an Cluster used to set Server
	// +optional
	ServerRef *xpv1.NamespacedReference `json:"serverRef,omitempty"`
	// SourceReposSelector selects references to Repositories used to set SourceRepos
	// +optional
	ServerSelector *xpv1.NamespacedSelector `json:"serverSelector,omitempty"`
	// Namespace specifies the target namespace for the application's resources.
	// The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Name is an alternate way of specifying the target cluster by its symbolic name
	// +optional
	Name *string `json:"name,omitempty"`
	// contains filtered or unexported fields
}

// ProjectRole represents a role that has access to a project
type ProjectRole struct {
	// Name is a name for this role
	Name string `json:"name"`
	// Description is a description of the role
	// +optional
	Description *string `json:"description,omitempty"`
	// Policies Stores a list of casbin formated strings that define access policies for the role in the project
	// +optional
	Policies []string `json:"policies,omitempty"`
	// JWTTokens are a list of generated JWT tokens bound to this role
	// +optional
	JWTTokens []JWTToken `json:"jwtTokens,omitempty"`
	// Groups are a list of OIDC group claims bound to this role
	// +optional
	Groups []string `json:"groups,omitempty"`
}

// JWTToken holds the issuedAt and expiresAt values of a token
type JWTToken struct {
	IssuedAt int64 `json:"iat"`
	// +optional
	ExpiresAt *int64 `json:"exp,omitempty"`
	// +optional
	ID *string `json:"id,omitempty"`
}

// JWTTokens represents a list of JWT tokens
type JWTTokens struct {
	// +optional
	Items []JWTToken `json:"items,omitempty"`
}

// OrphanedResourcesMonitorSettings holds settings of orphaned resources monitoring
type OrphanedResourcesMonitorSettings struct {
	// Warn indicates if warning condition should be created for apps which have orphaned resources
	// +optional
	Warn *bool `json:"warn,omitempty"`
	// Ignore contains a list of resources that are to be excluded from orphaned resources monitoring
	// +optional
	Ignore []OrphanedResourceKey `json:"ignore,omitempty"`
}

// OrphanedResourceKey is a reference to a resource to be ignored from
type OrphanedResourceKey struct {
	// +optional
	Group *string `json:"group,omitempty"`
	// +optional
	Kind *string `json:"kind,omitempty"`
	// +optional
	Name *string `json:"name,omitempty"`
}

// SyncWindows is a collection of sync windows in this project
type SyncWindows []SyncWindow

// SyncWindow contains the kind, time, duration and attributes that are used to assign the syncWindows to apps
type SyncWindow struct {
	// Kind defines if the window allows or blocks syncs
	// +optional
	Kind *string `json:"kind,omitempty"`
	// Schedule is the time the window will begin, specified in cron format
	// +optional
	Schedule *string `json:"schedule,omitempty"`
	// Duration is the amount of time the sync window will be open
	// +optional
	Duration *string `json:"duration,omitempty"`
	// Applications contains a list of applications that the window will apply to
	// +optional
	Applications []string `json:"applications,omitempty"`
	// Namespaces contains a list of namespaces that the window will apply to
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`
	// Clusters contains a list of clusters that the window will apply to
	// +optional
	Clusters []string `json:"clusters,omitempty"`
	// ManualSync enables manual syncs when they would otherwise be blocked
	// +optional
	ManualSync *bool `json:"manualSync,omitempty"`
}

// SignatureKey is the specification of a key required to verify commit signatures with
type SignatureKey struct {
	// The ID of the key in hexadecimal notation
	KeyID string `json:"keyID"`
}

// ProjectObservation represents an argocd Project.
type ProjectObservation struct {
	// JWTTokensByRole contains a list of JWT tokens issued for a given role
	// +optional
	JWTTokensByRole map[string]JWTTokens `json:"jwtTokensByRole,omitempty"`
}

// A ProjectSpec defines the desired state of an ArgoCD Project.
type ProjectSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ProjectParameters `json:"forProvider"`
}

// A ProjectStatus represents the observed state of an ArgoCD Project.
type ProjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Project is a managed resource that represents an ArgoCD Git Project
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project items
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}
