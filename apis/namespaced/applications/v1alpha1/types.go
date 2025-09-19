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
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ApplicationParameters define the desired state of an ArgoCD Git Application
type ApplicationParameters struct {
	Source *ApplicationSource `json:"source,omitempty" protobuf:"bytes,1,opt,name=source"`
	// Destination is a reference to the target Kubernetes server and namespace
	Destination ApplicationDestination `json:"destination" protobuf:"bytes,2,name=destination"`
	// Project is a reference to the project this application belongs to.
	// The empty string means that application belongs to the 'default' project.
	Project string `json:"project" protobuf:"bytes,3,name=project"`
	// SyncPolicy controls when and how a sync will be performed
	SyncPolicy *SyncPolicy `json:"syncPolicy,omitempty" protobuf:"bytes,4,name=syncPolicy"`
	// IgnoreDifferences is a list of resources and their fields which should be ignored during comparison
	IgnoreDifferences []ResourceIgnoreDifferences `json:"ignoreDifferences,omitempty" protobuf:"bytes,5,name=ignoreDifferences"`
	// Info contains a list of information (URLs, email addresses, and plain text) that relates to the application
	Info []Info `json:"info,omitempty" protobuf:"bytes,6,name=info"`
	// RevisionHistoryLimit limits the number of items kept in the application's revision history, which is used for informational purposes as well as for rollbacks to previous versions.
	// This should only be changed in exceptional circumstances.
	// Setting to zero will store no history. This will reduce storage used.
	// Increasing will increase the space used to store the history, so we do not recommend increasing it.
	// Default is 10.
	RevisionHistoryLimit *int64 `json:"revisionHistoryLimit,omitempty" protobuf:"bytes,7,name=revisionHistoryLimit"`

	// Sources is a reference to the location of the application's manifests or chart
	Sources ApplicationSources `json:"sources,omitempty" protobuf:"bytes,8,opt,name=sources"`

	// Annotations that will be applied to the ArgoCD Application
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,opt,name=annotations"`

	// Finalizers added to the ArgoCD Application
	Finalizers []string `json:"finalizers,omitempty" protobuf:"bytes,13,opt,name=finalizers"`

	// SourceHydrator provides a way to push hydrated manifests back to git before syncing them to the cluster.
	SourceHydrator *SourceHydrator `json:"sourceHydrator,omitempty" protobuf:"bytes,14,opt,name=sourceHydrator"`

	// AppNamespace is the namespace of the application in the ArgoCD server
	AppNamespace *string `json:"appNamespace,omitempty"`

	// DeleteCascade indicates whether the delete should be cascaded to the app's underlying resources
	DeleteCascade *bool `json:"deleteCascade,omitempty"`
	// DeletePropagationPolicy defines the policy for propagating deletions to the app's resources
	DeletePropagationPolicy *string `json:"deletePropagationPolicy,omitempty"`
}

// ResourceIgnoreDifferences contains resource filter and list of json paths which should be ignored during comparison with live state.
type ResourceIgnoreDifferences struct {
	Group             string   `json:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Kind              string   `json:"kind" protobuf:"bytes,2,opt,name=kind"`
	Name              string   `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
	Namespace         string   `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
	JSONPointers      []string `json:"jsonPointers,omitempty" protobuf:"bytes,5,opt,name=jsonPointers"`
	JQPathExpressions []string `json:"jqPathExpressions,omitempty" protobuf:"bytes,6,opt,name=jqPathExpressions"`
	// ManagedFieldsManagers is a list of trusted managers. Fields mutated by those managers will take precedence over the
	// desired state defined in the SCM and won't be displayed in diffs
	ManagedFieldsManagers []string `json:"managedFieldsManagers,omitempty" protobuf:"bytes,7,opt,name=managedFieldsManagers"`
}

// EnvEntry represents an entry in the application's environment
type EnvEntry struct {
	// Name is the name of the variable, usually expressed in uppercase
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Value is the value of the variable
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

// ApplicationSource contains all required information about the source of an application
type ApplicationSource struct {
	// RepoURL is the URL to the repository (Git or Helm) that contains the application manifests
	RepoURL string `json:"repoURL" protobuf:"bytes,1,opt,name=repoURL"`
	// Path is a directory path within the Git repository, and is only valid for applications sourced from Git.
	Path *string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`
	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD.
	// In case of Helm, this is a semver tag for the Chart's version.
	TargetRevision *string `json:"targetRevision,omitempty" protobuf:"bytes,4,opt,name=targetRevision"`
	// Helm holds helm specific options
	Helm *ApplicationSourceHelm `json:"helm,omitempty" protobuf:"bytes,7,opt,name=helm"`
	// Kustomize holds kustomize specific options
	Kustomize *ApplicationSourceKustomize `json:"kustomize,omitempty" protobuf:"bytes,8,opt,name=kustomize"`
	// Directory holds path/directory specific options
	Directory *ApplicationSourceDirectory `json:"directory,omitempty" protobuf:"bytes,10,opt,name=directory"`
	// Plugin holds config management plugin specific options
	Plugin *ApplicationSourcePlugin `json:"plugin,omitempty" protobuf:"bytes,11,opt,name=plugin"`
	// Chart is a Helm chart name, and must be specified for applications sourced from a Helm repo.
	Chart *string `json:"chart,omitempty" protobuf:"bytes,12,opt,name=chart"`
	// Ref is reference to another source within sources field. This field will not be used if used with a `source` tag.
	Ref *string `json:"ref,omitempty" protobuf:"bytes,13,opt,name=ref"`
	// Name is the name of the application source
	Name string `json:"name,omitempty" protobuf:"bytes,14,opt,name=name"`
}

// ApplicationSources contains list of required information about the sources of an application
type ApplicationSources []ApplicationSource

// SourceHydrator specifies a dry "don't repeat yourself" source for manifests, a sync source from which to sync
// hydrated manifests, and an optional hydrateTo location to act as a "staging" aread for hydrated manifests.
type SourceHydrator struct {
	// DrySource specifies where the dry "don't repeat yourself" manifest source lives.
	DrySource DrySource `json:"drySource" protobuf:"bytes,1,name=drySource"`
	// SyncSource specifies where to sync hydrated manifests from.
	SyncSource SyncSource `json:"syncSource" protobuf:"bytes,2,name=syncSource"`
	// HydrateTo specifies an optional "staging" location to push hydrated manifests to. An external system would then
	// have to move manifests to the SyncSource, e.g. by pull request.
	HydrateTo *HydrateTo `json:"hydrateTo,omitempty" protobuf:"bytes,3,opt,name=hydrateTo"`
}

// DrySource specifies a location for dry "don't repeat yourself" manifest source information.
type DrySource struct {
	// RepoURL is the URL to the git repository that contains the application manifests
	RepoURL string `json:"repoURL" protobuf:"bytes,1,name=repoURL"`
	// TargetRevision defines the revision of the source to hydrate
	TargetRevision string `json:"targetRevision" protobuf:"bytes,2,name=targetRevision"`
	// Path is a directory path within the Git repository where the manifests are located
	Path string `json:"path" protobuf:"bytes,3,name=path"`
}

// SyncSource specifies a location from which hydrated manifests may be synced. RepoURL is assumed based on the
// associated DrySource config in the SourceHydrator.
type SyncSource struct {
	// TargetBranch is the branch to which hydrated manifests should be committed
	TargetBranch string `json:"targetBranch" protobuf:"bytes,1,name=targetBranch"`
	// Path is a directory path within the git repository where hydrated manifests should be committed to and synced
	// from. If hydrateTo is set, this is just the path from which hydrated manifests will be synced.
	Path string `json:"path" protobuf:"bytes,2,name=path"`
}

// HydrateTo specifies a location to which hydrated manifests should be pushed as a "staging area" before being moved to
// the SyncSource. The RepoURL and Path are assumed based on the associated SyncSource config in the SourceHydrator.
type HydrateTo struct {
	// TargetBranch is the branch to which hydrated manifests should be committed
	TargetBranch string `json:"targetBranch" protobuf:"bytes,1,name=targetBranch"`
}

// ApplicationDestination holds information about the application's destination
type ApplicationDestination struct {
	// Server specifies the URL of the target cluster and must be set to the Kubernetes control plane API
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1.Cluster
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1.ServerAddress()
	// +crossplane:generate:reference:refFieldName=ServerRef
	// +crossplane:generate:reference:selectorFieldName=ServerSelector
	// +optional
	Server *string `json:"server,omitempty"`
	// ServerRef is a reference to Cluster used to set Server
	// +optional
	ServerRef *xpv1.Reference `json:"serverRef,omitempty"`
	// ServerSelector selects references to Cluster used to set Server
	// +optional
	ServerSelector *xpv1.Selector `json:"serverSelector,omitempty"`
	// Namespace specifies the target namespace for the application's resources.
	// The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Name is an alternate way of specifying the target cluster by its symbolic name
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1.Cluster
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-argocd/apis/cluster/v1alpha1.ServerName()
	// +crossplane:generate:reference:refFieldName=NameRef
	// +crossplane:generate:reference:selectorFieldName=NameSelector
	// +optional
	Name *string `json:"name,omitempty"`
	// NameRef is a reference to a Cluster used to set Name
	// +optional
	NameRef *xpv1.Reference `json:"nameRef,omitempty"`
	// NameSelector is a reference to a Cluster used to set Name
	// +optional
	NameSelector *xpv1.Selector `json:"nameSelector,omitempty"`
	// contains filtered or unexported fields
}

// ConnectionState is the observed state of the argocd repository
type ConnectionState struct {
	Status     string       `json:"status,omitempty"`
	Message    string       `json:"message,omitempty"`
	ModifiedAt *metav1.Time `json:"attemptedAt,omitempty"`
}

// A ApplicationSpec defines the desired state of an ArgoCD Application.
type ApplicationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationParameters `json:"forProvider"`
}

// A ApplicationStatus represents the observed state of an ArgoCD Application.
type ApplicationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ArgoApplicationStatus `json:"atProvider,omitempty"`
}

// ApplicationSourceHelm holds helm specific options
type ApplicationSourceHelm struct {
	// ValuesFiles is a list of Helm value files to use when generating a template
	ValueFiles []string `json:"valueFiles,omitempty" protobuf:"bytes,1,opt,name=valueFiles"`
	// Parameters is a list of Helm parameters which are passed to the helm template command upon manifest generation
	Parameters []HelmParameter `json:"parameters,omitempty" protobuf:"bytes,2,opt,name=parameters"`
	// ReleaseName is the Helm release name to use. If omitted it will use the application name
	ReleaseName *string `json:"releaseName,omitempty" protobuf:"bytes,3,opt,name=releaseName"`
	// Values specifies Helm values to be passed to helm template, typically defined as a block. ValuesObject takes precedence over Values, so use one or the other.
	Values *string `json:"values,omitempty" patchStrategy:"replace" protobuf:"bytes,4,opt,name=values"`
	// FileParameters are file parameters to the helm template
	FileParameters []HelmFileParameter `json:"fileParameters,omitempty" protobuf:"bytes,5,opt,name=fileParameters"`
	// Version is the Helm version to use for templating ("3")
	Version *string `json:"version,omitempty" protobuf:"bytes,6,opt,name=version"`
	// PassCredentials pass credentials to all domains (Helm's --pass-credentials)
	PassCredentials *bool `json:"passCredentials,omitempty" protobuf:"bytes,7,opt,name=passCredentials"`
	// IgnoreMissingValueFiles prevents helm template from failing when valueFiles do not exist locally by not appending them to helm template --values
	IgnoreMissingValueFiles *bool `json:"ignoreMissingValueFiles,omitempty" protobuf:"bytes,8,opt,name=ignoreMissingValueFiles"`
	// SkipCrds skips custom resource definition installation step (Helm's --skip-crds)
	SkipCrds *bool `json:"skipCrds,omitempty" protobuf:"bytes,9,opt,name=skipCrds"`
	// ValuesObject specifies Helm values to be passed to helm template, defined as a map. This takes precedence over Values.
	ValuesObject extv1.JSON `json:"valuesObject,omitempty" protobuf:"bytes,10,opt,name=valuesObject"`
	// Namespace is an optional namespace to template with. If left empty, defaults to the app's destination namespace.
	Namespace *string `json:"namespace,omitempty" protobuf:"bytes,11,opt,name=namespace"`
	// KubeVersion specifies the Kubernetes API version to pass to Helm when templating manifests. By default, Argo CD
	// uses the Kubernetes version of the target cluster.
	KubeVersion *string `json:"kubeVersion,omitempty" protobuf:"bytes,12,opt,name=kubeVersion"`
	// APIVersions specifies the Kubernetes resource API versions to pass to Helm when templating manifests. By default,
	// Argo CD uses the API versions of the target cluster. The format is [group/]version/kind.
	APIVersions []string `json:"apiVersions,omitempty" protobuf:"bytes,13,opt,name=apiVersions"`
	// SkipTests skips test manifest installation step (Helm's --skip-tests).
	SkipTests bool `json:"skipTests,omitempty" protobuf:"bytes,14,opt,name=skipTests"`
	// SkipSchemaValidation skips JSON schema validation (Helm's --skip-schema-validation)
	SkipSchemaValidation bool `json:"skipSchemaValidation,omitempty" protobuf:"bytes,15,opt,name=skipSchemaValidation"`
}

// HelmParameter is a parameter that's passed to helm template during manifest generation
type HelmParameter struct {
	// Name is the name of the Helm parameter
	Name *string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Value is the value for the Helm parameter
	Value *string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// ForceString determines whether to tell Helm to interpret booleans and numbers as strings
	ForceString *bool `json:"forceString,omitempty" protobuf:"bytes,3,opt,name=forceString"`
}

// HelmFileParameter is a file parameter that's passed to helm template during manifest generation
type HelmFileParameter struct {
	// Name is the name of the Helm parameter
	Name *string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Path is the path to the file containing the values for the Helm parameter
	Path *string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`
}

// +kubebuilder:object:root=true

// An Application is a managed resource that represents an ArgoCD Application
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
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

// ApplicationSourceKustomize holds options specific to an Application source specific to Kustomize
type ApplicationSourceKustomize struct {
	// NamePrefix is a prefix appended to resources for Kustomize apps
	NamePrefix *string `json:"namePrefix,omitempty" protobuf:"bytes,1,opt,name=namePrefix"`
	// NameSuffix is a suffix appended to resources for Kustomize apps
	NameSuffix *string `json:"nameSuffix,omitempty" protobuf:"bytes,2,opt,name=nameSuffix"`
	// Images is a list of Kustomize image override specifications
	Images KustomizeImages `json:"images,omitempty" protobuf:"bytes,3,opt,name=images"`
	// CommonLabels is a list of additional labels to add to rendered manifests
	CommonLabels map[string]string `json:"commonLabels,omitempty" protobuf:"bytes,4,opt,name=commonLabels"`
	// Version controls which version of Kustomize to use for rendering manifests
	Version *string `json:"version,omitempty" protobuf:"bytes,5,opt,name=version"`
	// CommonAnnotations is a list of additional annotations to add to rendered manifests
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty" protobuf:"bytes,6,opt,name=commonAnnotations"`
	// ForceCommonLabels specifies whether to force applying common labels to resources for Kustomize apps
	ForceCommonLabels *bool `json:"forceCommonLabels,omitempty" protobuf:"bytes,7,opt,name=forceCommonLabels"`
	// ForceCommonAnnotations specifies whether to force applying common annotations to resources for Kustomize apps
	ForceCommonAnnotations *bool `json:"forceCommonAnnotations,omitempty" protobuf:"bytes,8,opt,name=forceCommonAnnotations"`
	// Namespace sets the namespace that Kustomize adds to all resources
	Namespace *string `json:"namespace,omitempty" protobuf:"bytes,9,opt,name=namespace"`
	// CommonAnnotationsEnvsubst specifies whether to apply env variables substitution for annotation values
	CommonAnnotationsEnvsubst *bool `json:"commonAnnotationsEnvsubst,omitempty" protobuf:"bytes,10,opt,name=commonAnnotationsEnvsubst"`
	// Replicas is a list of Kustomize Replicas override specifications
	Replicas KustomizeReplicas `json:"replicas,omitempty" protobuf:"bytes,11,opt,name=replicas"`
	// Patches is a list of Kustomize patches
	Patches KustomizePatches `json:"patches,omitempty" protobuf:"bytes,12,opt,name=patches"`
	// Components specifies a list of kustomize components to add to the kustomization before building
	Components []string `json:"components,omitempty" protobuf:"bytes,13,rep,name=components"`
	// IgnoreMissingComponents prevents kustomize from failing when components do not exist locally by not appending them to kustomization file
	IgnoreMissingComponents bool `json:"ignoreMissingComponents,omitempty" protobuf:"bytes,17,opt,name=ignoreMissingComponents"`
	// LabelWithoutSelector specifies whether to apply common labels to resource selectors or not
	LabelWithoutSelector *bool `json:"labelWithoutSelector,omitempty" protobuf:"bytes,14,opt,name=labelWithoutSelector"`
	// KubeVersion specifies the Kubernetes API version to pass to Helm when templating manifests. By default, Argo CD
	// uses the Kubernetes version of the target cluster.
	KubeVersion *string `json:"kubeVersion,omitempty" protobuf:"bytes,15,opt,name=kubeVersion"`
	// APIVersions specifies the Kubernetes resource API versions to pass to Helm when templating manifests. By default,
	// Argo CD uses the API versions of the target cluster. The format is [group/]version/kind.
	APIVersions []string `json:"apiVersions,omitempty" protobuf:"bytes,16,opt,name=apiVersions"`
	// LabelIncludeTemplates specifies whether to apply common labels to resource templates or not
	LabelIncludeTemplates bool `json:"labelIncludeTemplates,omitempty" protobuf:"bytes,18,opt,name=labelIncludeTemplates"`
}

// KustomizeReplica override specifications
type KustomizeReplica struct {
	// Name of Deployment or StatefulSet
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// Number of replicas
	Count intstr.IntOrString `json:"count" protobuf:"bytes,2,name=count"`
}

// KustomizeReplicas is a list of KustomizeReplica override specifications
type KustomizeReplicas []KustomizeReplica

// KustomizeImages is a list of Kustomize images
type KustomizeImages []KustomizeImage

// KustomizeImage represents a Kustomize image definition in the format [old_image_name=]<image_name>:<image_tag>
type KustomizeImage string

// Info is a list of informational items for this operation
type Info struct {
	Name  string `json:"name" protobuf:"bytes,1,name=name"`
	Value string `json:"value" protobuf:"bytes,2,name=value"`
}

// KustomizePatches is a list of KustomizePatches
type KustomizePatches []KustomizePatch

// KustomizePatch is a kustomize patch
type KustomizePatch struct {
	Path    string             `json:"path,omitempty" yaml:"path,omitempty" protobuf:"bytes,1,opt,name=path"`
	Patch   string             `json:"patch,omitempty" yaml:"patch,omitempty" protobuf:"bytes,2,opt,name=patch"`
	Target  *KustomizeSelector `json:"target,omitempty" yaml:"target,omitempty" protobuf:"bytes,3,opt,name=target"`
	Options map[string]bool    `json:"options,omitempty" yaml:"options,omitempty" protobuf:"bytes,4,opt,name=options"`
}

// KustomizeSelector is a selector of a Kustomize Patch
type KustomizeSelector struct {
	KustomizeResId     `json:",inline,omitempty" yaml:",inline,omitempty" protobuf:"bytes,1,opt,name=resId"`
	AnnotationSelector string `json:"annotationSelector,omitempty" yaml:"annotationSelector,omitempty" protobuf:"bytes,2,opt,name=annotationSelector"`
	LabelSelector      string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty" protobuf:"bytes,3,opt,name=labelSelector"`
}

// KustomizeResId identifies a resource. Matches capitalization of upstream type
type KustomizeResId struct { //nolint:golint
	KustomizeGvk `json:",inline,omitempty" yaml:",inline,omitempty" protobuf:"bytes,1,opt,name=gvk"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	Namespace    string `json:"namespace,omitempty" yaml:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
}

// KustomizeGvk contains group/version/kind for a resource
type KustomizeGvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Version string `json:"version,omitempty" yaml:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
}

// SyncOptions provide per-sync sync-options, e.g. Validate=false
type SyncOptions []string

// SyncPolicy controls when a sync will be performed in response to updates in git
type SyncPolicy struct {
	// Automated will keep an application synced to the target revision
	Automated *SyncPolicyAutomated `json:"automated,omitempty" protobuf:"bytes,1,opt,name=automated"`
	// Options allow you to specify whole app sync-options
	SyncOptions SyncOptions `json:"syncOptions,omitempty" protobuf:"bytes,2,opt,name=syncOptions"`
	// Retry controls failed sync retry behavior
	Retry *RetryStrategy `json:"retry,omitempty" protobuf:"bytes,3,opt,name=retry"`
	// ManagedNamespaceMetadata controls metadata in the given namespace (if CreateNamespace=true)
	ManagedNamespaceMetadata *ManagedNamespaceMetadata `json:"managedNamespaceMetadata,omitempty" protobuf:"bytes,4,opt,name=managedNamespaceMetadata"`
}

// SyncPolicyAutomated controls the behavior of an automated sync
type SyncPolicyAutomated struct {
	// Prune specifies whether to delete resources from the cluster that are not found in the sources anymore as part of automated sync (default: false)
	Prune *bool `json:"prune,omitempty" protobuf:"bytes,1,opt,name=prune"`
	// SelfHeal specifes whether to revert resources back to their desired state upon modification in the cluster (default: false)
	SelfHeal *bool `json:"selfHeal,omitempty" protobuf:"bytes,2,opt,name=selfHeal"`
	// AllowEmpty allows apps have zero live resources (default: false)
	AllowEmpty *bool `json:"allowEmpty,omitempty" protobuf:"bytes,3,opt,name=allowEmpty"`
}

// SyncStrategy controls the manner in which a sync is performed
type SyncStrategy struct {
	// Apply will perform a `kubectl apply` to perform the sync.
	Apply *SyncStrategyApply `json:"apply,omitempty" protobuf:"bytes,1,opt,name=apply"`
	// Hook will submit any referenced resources to perform the sync. This is the default strategy
	Hook *SyncStrategyHook `json:"hook,omitempty" protobuf:"bytes,2,opt,name=hook"`
}

// Backoff is the backoff strategy to use on subsequent retries for failing syncs
type Backoff struct {
	// Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. "2m", "1h")
	Duration *string `json:"duration,omitempty" protobuf:"bytes,1,opt,name=duration"`
	// Factor is a factor to multiply the base duration after each failed retry
	Factor *int64 `json:"factor,omitempty" protobuf:"bytes,2,name=factor"`
	// MaxDuration is the maximum amount of time allowed for the backoff strategy
	MaxDuration *string `json:"maxDuration,omitempty" protobuf:"bytes,3,opt,name=maxDuration"`
}

// SyncStrategyApply uses `kubectl apply` to perform the apply
type SyncStrategyApply struct {
	// Force indicates whether or not to supply the --force flag to `kubectl apply`.
	// The --force flag deletes and re-create the resource, when PATCH encounters conflict and has
	// retried for 5 times.
	Force *bool `json:"force,omitempty" protobuf:"bytes,1,opt,name=force"`
}

// SyncStrategyHook will perform a sync using hooks annotations.
// If no hook annotation is specified falls back to `kubectl apply`.
type SyncStrategyHook struct {
	// Embed SyncStrategyApply type to inherit any `apply` options
	// +optional
	SyncStrategyApply `json:",inline" protobuf:"bytes,1,opt,name=syncStrategyApply"`
}

// RetryStrategy controls the strategy to apply if a sync fails
type RetryStrategy struct {
	// Limit is the maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.
	Limit *int64 `json:"limit,omitempty" protobuf:"bytes,1,opt,name=limit"`
	// Backoff controls how to backoff on subsequent retries of failed syncs
	Backoff *Backoff `json:"backoff,omitempty" protobuf:"bytes,2,opt,name=backoff,casttype=Backoff"`
}

// ManagedNamespaceMetadata controls metadata in the given namespace (if CreateNamespace=true)
type ManagedNamespaceMetadata struct {
	Labels      map[string]string `json:"labels,omitempty" protobuf:"bytes,1,opt,name=labels"`
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,2,opt,name=annotations"`
}

// ApplicationSourceDirectory holds config management plugin specific options
type ApplicationSourceDirectory struct {
	// Recurse specifies whether to scan a directory recursively for manifests
	Recurse *bool `json:"recurse,omitempty" protobuf:"bytes,1,opt,name=recurse"`
	// Jsonnet holds options specific to Jsonnet
	Jsonnet ApplicationSourceJsonnet `json:"jsonnet,omitempty" protobuf:"bytes,2,opt,name=jsonnet"`
	// Exclude contains a glob pattern to match paths against that should be explicitly excluded from being used during manifest generation
	Exclude *string `json:"exclude,omitempty" protobuf:"bytes,3,opt,name=exclude"`
	// Include contains a glob pattern to match paths against that should be explicitly included during manifest generation
	Include *string `json:"include,omitempty" protobuf:"bytes,4,opt,name=include"`
}

// ApplicationSourceJsonnet holds options specific to Jsonnet
type ApplicationSourceJsonnet struct {
	// ExtVars is a list of Jsonnet External Variables
	ExtVars []JsonnetVar `json:"extVars,omitempty" protobuf:"bytes,1,opt,name=extVars"`
	// TLAS is a list of Jsonnet Top-level Arguments
	TLAs []JsonnetVar `json:"tlas,omitempty" protobuf:"bytes,2,opt,name=tlas"`
	// Additional library search dirs
	Libs []string `json:"libs,omitempty" protobuf:"bytes,3,opt,name=libs"`
}

// JsonnetVar represents a variable to be passed to jsonnet during manifest generation
type JsonnetVar struct {
	Name  string `json:"name" protobuf:"bytes,1,opt,name=name"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
	Code  *bool  `json:"code,omitempty" protobuf:"bytes,3,opt,name=code"`
}

// ApplicationSourcePluginParameters is a list of specific config management parameters
type ApplicationSourcePluginParameters []ApplicationSourcePluginParameter

// ApplicationSourcePluginParameter holds options specific to config management parameters
type ApplicationSourcePluginParameter struct {
	// Name is the name identifying a parameter.
	Name *string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// String_ is the value of a string type parameter.
	String_ *string `json:"string,omitempty" protobuf:"bytes,5,opt,name=string"`
	// Map is the value of a map type parameter.
	*OptionalMap `json:",omitempty" protobuf:"bytes,3,rep,name=map"`
	// Array is the value of an array type parameter.
	*OptionalArray `json:",omitempty" protobuf:"bytes,4,rep,name=array"`
}

// OptionalMap is the value of a map type parameter.
type OptionalMap struct {
	// Map is the value of a map type parameter.
	// +optional
	Map map[string]string `json:"map" protobuf:"bytes,1,rep,name=map"`
	// We need the explicit +optional so that kube-builder generates the CRD without marking this as required.
}

// OptionalArray is the value of an array type parameter.
type OptionalArray struct {
	// Array is the value of an array type parameter.
	// +optional
	Array []string `json:"array" protobuf:"bytes,1,rep,name=array"`
	// We need the explicit +optional so that kube-builder generates the CRD without marking this as required.
}

// ApplicationSourcePlugin holds options specific to config management plugins
type ApplicationSourcePlugin struct {
	Name       *string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Env        `json:"env,omitempty" protobuf:"bytes,2,opt,name=env"`
	Parameters ApplicationSourcePluginParameters `json:"parameters,omitempty" protobuf:"bytes,3,opt,name=parameters"`
}

// Env holds options specific to config management plugins
type Env []*EnvEntry
