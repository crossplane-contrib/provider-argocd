package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoApplicationStatus contains status information for the application
type ArgoApplicationStatus struct {
	// Resources is a list of Kubernetes resources managed by this application
	Resources []ResourceStatus `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
	// Sync contains information about the application's current sync status
	Sync SyncStatus `json:"sync,omitempty" protobuf:"bytes,2,opt,name=sync"`
	// Health contains information about the application's current health status
	Health HealthStatus `json:"health,omitempty" protobuf:"bytes,3,opt,name=health"`
	// History contains information about the application's sync history
	History RevisionHistories `json:"history,omitempty" protobuf:"bytes,4,opt,name=history"`
	// Conditions is a list of currently observed application conditions
	Conditions []ApplicationCondition `json:"conditions,omitempty" protobuf:"bytes,5,opt,name=conditions"`
	// ReconciledAt indicates when the application state was reconciled using the latest git version
	ReconciledAt *metav1.Time `json:"reconciledAt,omitempty" protobuf:"bytes,6,opt,name=reconciledAt"`
	// OperationState contains information about any ongoing operations, such as a sync
	OperationState *OperationState `json:"operationState,omitempty" protobuf:"bytes,7,opt,name=operationState"`
	// ObservedAt indicates when the application state was updated without querying latest git state
	// Deprecated: controller no longer updates ObservedAt field
	ObservedAt *metav1.Time `json:"observedAt,omitempty" protobuf:"bytes,8,opt,name=observedAt"`
	// SourceType specifies the type of this application
	SourceType ApplicationSourceType `json:"sourceType,omitempty" protobuf:"bytes,9,opt,name=sourceType"`
	// Summary contains a list of URLs and container images used by this application
	Summary ApplicationSummary `json:"summary,omitempty" protobuf:"bytes,10,opt,name=summary"`
	// ResourceHealthSource indicates where the resource health status is stored: inline if not set or appTree
	ResourceHealthSource string `json:"resourceHealthSource,omitempty" protobuf:"bytes,11,opt,name=resourceHealthSource"`
	// SourceTypes specifies the type of the sources included in the application
	SourceTypes []ApplicationSourceType `json:"sourceTypes,omitempty" protobuf:"bytes,12,opt,name=sourceTypes"`
}

// RevisionHistories contains information about the application's sync history
type RevisionHistories []RevisionHistory

// RevisionHistory contains history information about a previous sync
type RevisionHistory struct {
	// Revision holds the revision the sync was performed against
	Revision *string `json:"revision,omitempty" protobuf:"bytes,2,opt,name=revision"`
	// DeployedAt holds the time the sync operation completed
	DeployedAt *metav1.Time `json:"deployedAt,omitempty" protobuf:"bytes,4,opt,name=deployedAt"`
	// ID is an auto incrementing identifier of the RevisionHistory
	ID *int64 `json:"id" protobuf:"bytes,5,opt,name=id"`
	// Source is a reference to the application source used for the sync operation
	Source ApplicationSource `json:"source,omitempty" protobuf:"bytes,6,opt,name=source"`
	// DeployStartedAt holds the time the sync operation started
	DeployStartedAt *metav1.Time `json:"deployStartedAt,omitempty" protobuf:"bytes,7,opt,name=deployStartedAt"`
	// Sources is a reference to the application sources used for the sync operation
	Sources ApplicationSources `json:"sources,omitempty" protobuf:"bytes,8,opt,name=sources"`
	// Revisions holds the revision of each source in sources field the sync was performed against
	Revisions []string `json:"revisions,omitempty" protobuf:"bytes,9,opt,name=revisions"`
}

// ResourceStatus holds the current sync and health status of a resource
type ResourceStatus struct {
	Group                        *string       `json:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Version                      *string       `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Kind                         *string       `json:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
	Namespace                    *string       `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
	Name                         *string       `json:"name,omitempty" protobuf:"bytes,5,opt,name=name"`
	Status                       *string       `json:"status,omitempty" protobuf:"bytes,6,opt,name=status"`
	Health                       *HealthStatus `json:"health,omitempty" protobuf:"bytes,7,opt,name=health"`
	Hook                         *bool         `json:"hook,omitempty" protobuf:"bytes,8,opt,name=hook"`
	RequiresPruning              *bool         `json:"requiresPruning,omitempty" protobuf:"bytes,9,opt,name=requiresPruning"`
	SyncWave                     *int64        `json:"syncWave,omitempty" protobuf:"bytes,10,opt,name=syncWave"`
	RequiresDeletionConfirmation *bool         `json:"requiresDeletionConfirmation,omitempty" protobuf:"bytes,11,opt,name=requiresDeletionConfirmation"`
}

// SyncStatus contains information about the currently observed live and desired states of an application
type SyncStatus struct {
	// Status is the sync state of the comparison
	Status string `json:"status" protobuf:"bytes,1,opt,name=status,casttype=SyncStatusCode"`
	// ComparedTo contains information about what has been compared
	ComparedTo ComparedTo `json:"comparedTo,omitempty" protobuf:"bytes,2,opt,name=comparedTo"`
	// Revision contains information about the revision the comparison has been performed to
	Revision *string `json:"revision,omitempty" protobuf:"bytes,3,opt,name=revision"`
	// Revisions contains information about the revisions of multiple sources the comparison has been performed to
	Revisions []string `json:"revisions,omitempty" protobuf:"bytes,4,opt,name=revisions"`
}

// HealthStatus contains information about the currently observed health state of an application or resource
type HealthStatus struct {
	// Status holds the status code of the application or resource
	Status string `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	// Message is a human-readable informational message describing the health status
	Message *string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// LastTransitionTime is the time the HealthStatus was set or updated
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
}

// InfoItem contains arbitrary, human readable information about an application
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name *string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Value is human readable content.
	Value *string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

// ComparedTo contains application source and target which was used for resources comparison
type ComparedTo struct {
	// Source is a reference to the application's source used for comparison
	Source ApplicationSource `json:"source,omitempty" protobuf:"bytes,1,opt,name=source"`
	// Destination is a reference to the application's destination used for comparison
	Destination ApplicationDestination `json:"destination" protobuf:"bytes,2,opt,name=destination"`
	// Sources is a reference to the application's multiple sources used for comparison
	Sources ApplicationSources `json:"sources,omitempty" protobuf:"bytes,3,opt,name=sources"`
}

// ApplicationCondition contains details about an application condition, which is usually an error or warning
type ApplicationCondition struct {
	// Type is an application condition type
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Message contains human-readable message indicating details about condition
	Message string `json:"message" protobuf:"bytes,2,opt,name=message"`
	// LastTransitionTime is the time the condition was last observed
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
}

// OperationState contains information about state of a running operation
type OperationState struct {
	// Operation is the original requested operation
	Operation Operation `json:"operation,omitempty" protobuf:"bytes,1,opt,name=operation"`
	// Phase is the current phase of the operation
	Phase OperationPhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase"`
	// Message holds any pertinent messages when attempting to perform operation (typically errors).
	Message *string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// SyncResult is the result of a Sync operation
	SyncResult *SyncOperationResult `json:"syncResult,omitempty" protobuf:"bytes,4,opt,name=syncResult"`
	// StartedAt contains time of operation start
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,6,opt,name=startedAt"`
	// FinishedAt contains time of operation completion
	FinishedAt *metav1.Time `json:"finishedAt,omitempty" protobuf:"bytes,7,opt,name=finishedAt"`
	// RetryCount contains time of operation retries
	RetryCount *int64 `json:"retryCount,omitempty" protobuf:"bytes,8,opt,name=retryCount"`
}

// OperationPhase specifies the phase of the sync
type OperationPhase string

// ApplicationSourceType specifies the type of the application's source
type ApplicationSourceType string

// ApplicationSummary contains information about URLs and container images used by an application
type ApplicationSummary struct {
	// ExternalURLs holds all external URLs of application child resources.
	ExternalURLs []string `json:"externalURLs,omitempty" protobuf:"bytes,1,opt,name=externalURLs"`
	// Images holds all images of application child resources.
	Images []string `json:"images,omitempty" protobuf:"bytes,2,opt,name=images"`
}

// OperationInitiator contains information about the initiator of an operation
type OperationInitiator struct {
	// Username contains the name of a user who started operation
	Username *string `json:"username,omitempty" protobuf:"bytes,1,opt,name=username"`
	// Automated is set to true if operation was initiated automatically by the application controller.
	Automated *bool `json:"automated,omitempty" protobuf:"bytes,2,opt,name=automated"`
}

// Operation contains information about a requested or running operation
type Operation struct {
	// Sync contains parameters for the operation
	Sync *SyncOperation `json:"sync,omitempty" protobuf:"bytes,1,opt,name=sync"`
	// InitiatedBy contains information about who initiated the operations
	InitiatedBy OperationInitiator `json:"initiatedBy,omitempty" protobuf:"bytes,2,opt,name=initiatedBy"`
	// Info is a list of informational items for this operation
	Info []*Info `json:"info,omitempty" protobuf:"bytes,3,name=info"`
	// Retry controls the strategy to apply if a sync fails
	Retry RetryStrategy `json:"retry,omitempty" protobuf:"bytes,4,opt,name=retry"`
}

// SyncOperationResult represent result of sync operation
type SyncOperationResult struct {
	// Resources contains a list of sync result items for each individual resource in a sync operation
	Resources ResourceResults `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
	// Revision holds the revision this sync operation was performed to
	Revision string `json:"revision" protobuf:"bytes,2,opt,name=revision"`
	// Source records the application source information of the sync, used for comparing auto-sync
	Source ApplicationSource `json:"source,omitempty" protobuf:"bytes,3,opt,name=source"`
	// Source records the application source information of the sync, used for comparing auto-sync
	Sources ApplicationSources `json:"sources,omitempty" protobuf:"bytes,4,opt,name=sources"`
	// Revisions holds the revision this sync operation was performed for respective indexed source in sources field
	Revisions []string `json:"revisions,omitempty" protobuf:"bytes,5,opt,name=revisions"`
}

// ResourceResults returns sync results
type ResourceResults []*ResourceResult

// ResourceResult holds the operation result details of a specific resource
type ResourceResult struct {
	// Group specifies the API group of the resource
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Version specifies the API version of the resource
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Kind specifies the API kind of the resource
	Kind string `json:"kind" protobuf:"bytes,3,opt,name=kind"`
	// Namespace specifies the target namespace of the resource
	Namespace string `json:"namespace" protobuf:"bytes,4,opt,name=namespace"`
	// Name specifies the name of the resource
	Name string `json:"name" protobuf:"bytes,5,opt,name=name"`
	// Status holds the final result of the sync. Will be empty if the resources is yet to be applied/pruned and is always zero-value for hooks
	Status *string `json:"status,omitempty" protobuf:"bytes,6,opt,name=status"`
	// Message contains an informational or error message for the last sync OR operation
	Message *string `json:"message,omitempty" protobuf:"bytes,7,opt,name=message"`
	// HookType specifies the type of the hook. Empty for non-hook resources
	HookType *string `json:"hookType,omitempty" protobuf:"bytes,8,opt,name=hookType"`
	// HookPhase contains the state of any operation associated with this resource OR hook
	// This can also contain values for non-hook resources.
	HookPhase OperationPhase `json:"hookPhase,omitempty" protobuf:"bytes,9,opt,name=hookPhase"`
	// SyncPhase indicates the particular phase of the sync that this result was acquired in
	SyncPhase *string `json:"syncPhase,omitempty" protobuf:"bytes,10,opt,name=syncPhase"`
}

// SyncOperation contains details about a sync operation.
type SyncOperation struct {
	// Revision is the revision (Git) or chart version (Helm) which to sync the application to
	// If omitted, will use the revision specified in app spec.
	Revision *string `json:"revision,omitempty" protobuf:"bytes,1,opt,name=revision"`
	// Prune specifies to delete resources from the cluster that are no longer tracked in git
	Prune *bool `json:"prune,omitempty" protobuf:"bytes,2,opt,name=prune"`
	// DryRun specifies to perform a `kubectl apply --dry-run` without actually performing the sync
	DryRun *bool `json:"dryRun,omitempty" protobuf:"bytes,3,opt,name=dryRun"`
	// SyncStrategy describes how to perform the sync
	SyncStrategy *SyncStrategy `json:"syncStrategy,omitempty" protobuf:"bytes,4,opt,name=syncStrategy"`
	// Resources describes which resources shall be part of the sync
	Resources []SyncOperationResource `json:"resources,omitempty" protobuf:"bytes,6,opt,name=resources"`
	// Source overrides the source definition set in the application.
	// This is typically set in a Rollback operation and is nil during a Sync operation
	Source *ApplicationSource `json:"source,omitempty" protobuf:"bytes,7,opt,name=source"`
	// Manifests is an optional field that overrides sync source with a local directory for development
	Manifests []string `json:"manifests,omitempty" protobuf:"bytes,8,opt,name=manifests"`
	// SyncOptions provide per-sync sync-options, e.g. Validate=false
	SyncOptions SyncOptions `json:"syncOptions,omitempty" protobuf:"bytes,9,opt,name=syncOptions"`
	// Sources overrides the source definition set in the application.
	// This is typically set in a Rollback operation and is nil during a Sync operation
	Sources ApplicationSources `json:"sources,omitempty" protobuf:"bytes,10,opt,name=sources"`
	// Revisions is the list of revision (Git) or chart version (Helm) which to sync each source in sources field for the application to
	// If omitted, will use the revision specified in app spec.
	Revisions []string `json:"revisions,omitempty" protobuf:"bytes,11,opt,name=revisions"`
}

// SyncOperationResource contains resources to sync.
type SyncOperationResource struct {
	Group     *string `json:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Kind      string  `json:"kind" protobuf:"bytes,2,opt,name=kind"`
	Name      string  `json:"name" protobuf:"bytes,3,opt,name=name"`
	Namespace *string `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
}
