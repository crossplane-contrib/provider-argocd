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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ApplicationParameters define the desired state of an ArgoCD Application
type ApplicationParameters struct {
	// Source is a reference to the location ksonnet application definition
	Source ApplicationSource `json:"source"`
	// Destination overrides the kubernetes server and namespace defined in the environment ksonnet app.yaml
	Destination ApplicationDestination `json:"destination"`
	// Project is a application project name. Empty name means that application belongs to 'default' project.
	// +optional
	Project *string `json:"project"`
	// SyncPolicy controls when a sync will be performed
	// +optional
	SyncPolicy *SyncPolicy `json:"syncPolicy,omitempty"`
	// IgnoreDifferences controls resources fields which should be ignored during comparison
	// +optional
	IgnoreDifferences []ResourceIgnoreDifferences `json:"ignoreDifferences,omitempty"`
	// Infos contains a list of useful information (URLs, email addresses, and plain text) that relates to the application
	// +optional
	Info []Info `json:"info,omitempty"`
	// This limits this number of items kept in the apps revision history.
	// This should only be changed in exceptional circumstances.
	// Setting to zero will store no history. This will reduce storage used.
	// Increasing will increase the space used to store the history, so we do not recommend increasing it.
	// Default is 10.
	// +optional
	RevisionHistoryLimit *int64 `json:"revisionHistoryLimit,omitempty"`
}

type ApplicationSource struct {
	// RepoURL is the repository URL of the application manifests
	RepoURL string `json:"repoURL"`
	// Path is a directory path within the Git repository
	// +optional
	Path *string `json:"path,omitempty"`
	// TargetRevision defines the commit, tag, or branch in which to sync the application to.
	// If omitted, will sync to HEAD
	// +optional
	TargetRevision *string `json:"targetRevision,omitempty"`
	// Helm holds helm specific options
	// +optional
	Helm *ApplicationSourceHelm `json:"helm,omitempty"`
	// Kustomize holds kustomize specific options
	// +optional
	Kustomize *ApplicationSourceKustomize `json:"kustomize,omitempty"`
	// Ksonnet holds ksonnet specific options
	// +optional
	Ksonnet *ApplicationSourceKsonnet `json:"ksonnet,omitempty"`
	// Directory holds path/directory specific options
	// +optional
	Directory *ApplicationSourceDirectory `json:"directory,omitempty"`
	// ConfigManagementPlugin holds config management plugin specific options
	// +optional
	Plugin *ApplicationSourcePlugin `json:"plugin,omitempty"`
	// Chart is a Helm chart name
	// +optional
	Chart *string `json:"chart,omitempty"`
}

type ApplicationSourceHelm struct {
	// ValuesFiles is a list of Helm value files to use when generating a template
	// +optional
	ValueFiles []string `json:"valueFiles,omitempty"`
	// Parameters are parameters to the helm template
	// +optional
	Parameters []HelmParameter `json:"parameters,omitempty"`
	// The Helm release name. If omitted it will use the application name
	// +optional
	ReleaseName *string `json:"releaseName,omitempty"`
	// Values is Helm values, typically defined as a block
	// +optional
	Values *string `json:"values,omitempty"`
	// FileParameters are file parameters to the helm template
	// +optional
	FileParameters []HelmFileParameter `json:"fileParameters,omitempty"`
	// Version is the Helm version to use for templating with
	// +optional
	Version *string `json:"version,omitempty"`
}

type HelmFileParameter struct {
	// Name is the name of the helm parameter
	// +optional
	Name *string `json:"name,omitempty"`
	// Path is the path value for the helm parameter
	// +optional
	Path *string `json:"path,omitempty"`
}

type HelmParameter struct {
	// Name is the name of the helm parameter
	// +optional
	Name *string `json:"name,omitempty"`
	// Value is the value for the helm parameter
	// +optional
	Value *string `json:"value,omitempty"`
	// ForceString determines whether to tell Helm to interpret booleans and numbers as strings
	// +optional
	ForceString *bool `json:"forceString,omitempty"`
}

type ApplicationSourceKustomize struct {
	// NamePrefix is a prefix appended to resources for kustomize apps
	// +optional
	NamePrefix *string `json:"namePrefix,omitempty"`
	// NameSuffix is a suffix appended to resources for kustomize apps
	// +optional
	NameSuffix *string `json:"nameSuffix,omitempty"`
	// Images are kustomize image overrides
	// +optional
	Images KustomizeImages `json:"images,omitempty"`
	// CommonLabels adds additional kustomize commonLabels
	// +optional
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
	// Version contains optional Kustomize version
	// +optional
	Version *string `json:"version,omitempty"`
	// CommonAnnotations adds additional kustomize commonAnnotations
	// +optional
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
}
type KustomizeImages []KustomizeImage

type KustomizeImage string

type ApplicationSourceKsonnet struct {
	// Environment is a ksonnet application environment name
	// +optional
	Environment *string `json:"environment,omitempty"`
	// Parameters are a list of ksonnet component parameter override values
	// +optional
	Parameters []KsonnetParameter `json:"parameters,omitempty"`
}

type KsonnetParameter struct {
	// +optional
	Component *string `json:"component,omitempty"`
	// +optional
	Name *string `json:"name"`
	// +optional
	Value *string `json:"value"`
}

type ApplicationSourceDirectory struct {
	// +optional
	Recurse *bool `json:"recurse,omitempty"`
	// +optional
	Jsonnet ApplicationSourceJsonnet `json:"jsonnet,omitempty"`
	// +optional
	Exclude *string `json:"exclude,omitempty"`
}

type ApplicationSourceJsonnet struct {
	// ExtVars is a list of Jsonnet External Variables
	// +optional
	ExtVars []JsonnetVar `json:"extVars,omitempty"`
	// TLAS is a list of Jsonnet Top-level Arguments
	// +optional
	TLAs []JsonnetVar `json:"tlas,omitempty"`
	// Additional library search dirs
	// +optional
	Libs []string `json:"libs,omitempty"`
}

type JsonnetVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	// +optional
	Code bool `json:"code,omitempty"`
}

type ApplicationSourcePlugin struct {
	// +optional
	Name *string `json:"name,omitempty"`
	// +optional
	Env `json:"env,omitempty"`
}

type Env []*EnvEntry

type EnvEntry struct {
	// the name, usually uppercase
	Name string `json:"name"`
	// the value
	Value string `json:"value"`
}

type ApplicationDestination struct {
	// Server overrides the environment server value in the ksonnet app.yaml
	// +optional
	Server *string `json:"server,omitempty"`
	// Namespace overrides the environment namespace value in the ksonnet app.yaml
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Name of the destination cluster which can be used instead of server (url) field
	// +optional
	Name *string `json:"name,omitempty"`
}

type SyncPolicy struct {
	// Automated will keep an application synced to the target revision
	// +optional
	Automated *SyncPolicyAutomated `json:"automated,omitempty"`
	// Options allow you to specify whole app sync-options
	// +optional
	SyncOptions []*string `json:"syncOptions,omitempty"`
	// Retry controls failed sync retry behavior
	// +optional
	Retry *RetryStrategy `json:"retry,omitempty"`
}

type SyncPolicyAutomated struct {
	// Prune will prune resources automatically as part of automated sync (default: false)
	// +optional
	Prune *bool `json:"prune,omitempty"`
	// SelfHeal enables auto-syncing if  (default: false)
	// +optional
	SelfHeal *bool `json:"selfHeal,omitempty"`
	// AllowEmpty allows apps have zero live resources (default: false)
	// +optional
	AllowEmpty *bool `json:"allowEmpty,omitempty"`
}

// type SyncOptions []string

type RetryStrategy struct {
	// Limit is the maximum number of attempts when retrying a container
	// +optional
	Limit *int64 `json:"limit,omitempty"`

	// Backoff is a backoff strategy
	// +optional
	Backoff *Backoff `json:"backoff,omitempty"`
}

type Backoff struct {
	// Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. "2m", "1h")
	// +optional
	Duration *string `json:"duration,omitempty"`
	// Factor is a factor to multiply the base duration after each failed retry
	// +optional
	Factor *int64 `json:"factor,omitempty"`
	// MaxDuration is the maximum amount of time allowed for the backoff strategy
	// +optional
	MaxDuration *string `json:"maxDuration,omitempty"`
}

type ResourceIgnoreDifferences struct {
	// +optional
	Group *string `json:"group,omitempty"`
	Kind  string  `json:"kind"`
	// +optional
	Name *string `json:"name,omitempty"`
	// +optional
	Namespace    *string  `json:"namespace,omitempty"`
	JSONPointers []string `json:"jsonPointers"`
}

type Info struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ApplicationObservation represents the observed status at argocd.
type ApplicationObservation struct {
	// +optional
	Resources []*ResourceStatus `json:"resources,omitempty"`
	// +optional
	Sync *SyncStatus `json:"sync,omitempty"`
	// +optional
	Health *HealthStatus `json:"health,omitempty"`
	// +optional
	History RevisionHistories `json:"history,omitempty"`
	// +optional
	Conditions []ApplicationCondition `json:"conditions,omitempty"`
	// ReconciledAt indicates when the application state was reconciled using the latest git version
	// +optional
	ReconciledAt *metav1.Time `json:"reconciledAt,omitempty"`
	// +optional
	OperationState *OperationState `json:"operationState,omitempty"`
	// ObservedAt indicates when the application state was updated without querying latest git state
	// Deprecated: controller no longer updates ObservedAt field
	// +optional
	ObservedAt *metav1.Time `json:"observedAt,omitempty"`
	// +optional
	SourceType *string `json:"sourceType,omitempty"`
	// +optional
	Summary *ApplicationSummary `json:"summary,omitempty"`
}

type ResourceStatus struct {
	// +optional
	Group *string `json:"group,omitempty"`
	// +optional
	Version *string `json:"version,omitempty"`
	// +optional
	Kind *string `json:"kind,omitempty"`
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// +optional
	Name *string `json:"name,omitempty"`
	// +optional
	Status *string `json:"status,omitempty"`
	// +optional
	Health *HealthStatus `json:"health,omitempty"`
	// +optional
	Hook *bool `json:"hook,omitempty"`
	// +optional
	RequiresPruning *bool `json:"requiresPruning,omitempty"`
}

// type SyncStatusCode string

type HealthStatus struct {
	// +optional
	Status *string `json:"status,omitempty"`
	// +optional
	Message *string `json:"message,omitempty"`
}

type SyncStatus struct {
	Status *string `json:"status"`
	// +optional
	ComparedTo *ComparedTo `json:"comparedTo,omitempty"`
	// +optional
	Revision *string `json:"revision,omitempty"`
}

type ComparedTo struct {
	Source      ApplicationSource      `json:"source"`
	Destination ApplicationDestination `json:"destination"`
}

// type HealthStatusCode string

type RevisionHistories []RevisionHistory

type RevisionHistory struct {
	// Revision holds the revision of the sync
	Revision string `json:"revision"`
	// DeployedAt holds the time the deployment completed
	DeployedAt metav1.Time `json:"deployedAt"`
	// ID is an auto incrementing identifier of the RevisionHistory
	ID int64 `json:"id"`
	// +optional
	Source *ApplicationSource `json:"source,omitempty"`
	// DeployStartedAt holds the time the deployment started
	// +optional
	DeployStartedAt *metav1.Time `json:"deployStartedAt,omitempty"`
}

type ApplicationCondition struct {
	// Type is an application condition type
	Type ApplicationConditionType `json:"type"`
	// Message contains human-readable message indicating details about condition
	Message string `json:"message"`
	// LastTransitionTime is the time the condition was first observed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

type ApplicationConditionType string

type OperationState struct {
	// Operation is the original requested operation
	Operation Operation `json:"operation"`
	// Phase is the current phase of the operation
	Phase OperationPhase `json:"phase"`
	// Message hold any pertinent messages when attempting to perform operation (typically errors).
	// +optional
	Message *string `json:"message,omitempty"`
	// SyncResult is the result of a Sync operation
	// +optional
	SyncResult *SyncOperationResult `json:"syncResult,omitempty"`
	// StartedAt contains time of operation start
	StartedAt metav1.Time `json:"startedAt"`
	// FinishedAt contains time of operation completion
	// +optional
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`
	// RetryCount contains time of operation retries
	// +optional
	RetryCount *int64 `json:"retryCount,omitempty"`
}
type Operation struct {
	// +optional
	Sync *SyncOperation `json:"sync,omitempty"`
	// +optional
	InitiatedBy *OperationInitiator `json:"initiatedBy,omitempty"`
	// +optional
	Info []*Info `json:"info,omitempty"`
	// Retry controls failed sync retry behavior
	// +optional
	Retry *RetryStrategy `json:"retry,omitempty"`
}

type SyncOperation struct {
	// Revision is the revision in which to sync the application to.
	// If omitted, will use the revision specified in app spec.
	// +optional
	Revision *string `json:"revision,omitempty"`
	// Prune deletes resources that are no longer tracked in git
	// +optional
	Prune *bool `json:"prune,omitempty"`
	// DryRun will perform a `kubectl apply --dry-run` without actually performing the sync
	// +optional
	DryRun *bool `json:"dryRun,omitempty"`
	// SyncStrategy describes how to perform the sync
	// +optional
	SyncStrategy *SyncStrategy `json:"syncStrategy,omitempty"`
	// Resources describes which resources to sync
	// +optional
	Resources []SyncOperationResource `json:"resources,omitempty"`
	// Source overrides the source definition set in the application.
	// This is typically set in a Rollback operation and nil during a Sync operation
	// +optional
	Source *ApplicationSource `json:"source,omitempty"`
	// Manifests is an optional field that overrides sync source with a local directory for development
	// +optional
	Manifests []string `json:"manifests,omitempty"`
	// SyncOptions provide per-sync sync-options, e.g. Validate=false
	// +optional
	SyncOptions []*string `json:"syncOptions,omitempty"`
}

type SyncStrategy struct {
	// Apply will perform a `kubectl apply` to perform the sync.
	// +optional
	Apply *SyncStrategyApply `json:"apply,omitempty"`
	// Hook will submit any referenced resources to perform the sync. This is the default strategy
	// +optional
	Hook *SyncStrategyHook `json:"hook,omitempty"`
}

type SyncStrategyApply struct {
	// Force indicates whether or not to supply the --force flag to `kubectl apply`.
	// The --force flag deletes and re-create the resource, when PATCH encounters conflict and has
	// retried for 5 times.
	// +optional
	Force *bool `json:"force,omitempty"`
}

type SyncStrategyHook struct {
	// Embed SyncStrategyApply type to inherit any `apply` options
	// +optional
	SyncStrategyApply `json:"syncStrategyApply,inline"`
}

type SyncOperationResource struct {
	// +optional
	Group *string `json:"group,omitempty"`
	Kind  string  `json:"kind"`
	Name  string  `json:"name"`
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

type OperationInitiator struct {
	// Name of a user who started operation.
	// +optional
	Username *string `json:"username,omitempty"`
	// Automated is set to true if operation was initiated automatically by the application controller.
	// +optional
	Automated *bool `json:"automated,omitempty"`
}

type OperationPhase string

type SyncOperationResult struct {
	// Resources holds the sync result of each individual resource
	// +optional
	Resources ResourceResults `json:"resources,omitempty"`
	// Revision holds the revision of the sync
	Revision string `json:"revision"`
	// Source records the application source information of the sync, used for comparing auto-sync
	// +optional
	Source *ApplicationSource `json:"source,omitempty"`
}

type ResourceResults []*ResourceResult

type ResourceResult struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	// the final result of the sync, this is be empty if the resources is yet to be applied/pruned and is always zero-value for hooks
	// +optional
	Status *ResultCode `json:"status,omitempty"`
	// message for the last sync OR operation
	// +optional
	Message *string `json:"message,omitempty"`
	// the type of the hook, empty for non-hook resources
	// +optional
	HookType *HookType `json:"hookType,omitempty"`
	// the state of any operation associated with this resource OR hook
	// note: can contain values for non-hook resources
	// +optional
	HookPhase *OperationPhase `json:"hookPhase,omitempty"`
	// indicates the particular phase of the sync that this is for
	// +optional
	SyncPhase *SyncPhase `json:"syncPhase,omitempty"`
}

type ResultCode string
type HookType string
type SyncPhase string

// type ApplicationSourceType string

type ApplicationSummary struct {
	// ExternalURLs holds all external URLs of application child resources.
	// +optional
	ExternalURLs []string `json:"externalURLs,omitempty"`
	// Images holds all images of application child resources.
	// +optional
	Images []string `json:"images,omitempty"`
}

// A ApplicationSpec defines the desired state of an ArgoCD Application.
type ApplicationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationParameters `json:"forProvider"`
}

// A ApplicationStatus represents the observed state of an ArgoCD Application.
type ApplicationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ApplicationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Application is a managed resource that represents an ArgoCD Application
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,argocd}
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application items
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}
