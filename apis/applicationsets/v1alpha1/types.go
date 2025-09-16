/*
Copyright 2022 The Crossplane Authors.

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
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ApplicationSetParameters are the configurable fields of a ApplicationSet.
type ApplicationSetParameters struct {
	GoTemplate        bool                        `json:"goTemplate,omitempty" protobuf:"bytes,1,name=goTemplate"`
	Generators        []ApplicationSetGenerator   `json:"generators" protobuf:"bytes,2,name=generators"`
	Template          ApplicationSetTemplate      `json:"template" protobuf:"bytes,3,name=template"`
	SyncPolicy        *ApplicationSetSyncPolicy   `json:"syncPolicy,omitempty" protobuf:"bytes,4,name=syncPolicy"`
	Strategy          *ApplicationSetStrategy     `json:"strategy,omitempty" protobuf:"bytes,5,opt,name=strategy"`
	PreservedFields   *ApplicationPreservedFields `json:"preservedFields,omitempty" protobuf:"bytes,6,opt,name=preservedFields"`
	GoTemplateOptions []string                    `json:"goTemplateOptions,omitempty" protobuf:"bytes,7,opt,name=goTemplateOptions"`
	// ApplyNestedSelectors enables selectors defined within the generators of two level-nested matrix or merge generators
	ApplyNestedSelectors         bool                            `json:"applyNestedSelectors,omitempty" protobuf:"bytes,8,name=applyNestedSelectors"`
	IgnoreApplicationDifferences ApplicationSetIgnoreDifferences `json:"ignoreApplicationDifferences,omitempty" protobuf:"bytes,9,name=ignoreApplicationDifferences"`
	TemplatePatch                *string                         `json:"templatePatch,omitempty" protobuf:"bytes,10,name=templatePatch"`

	// AppsetNamespace is the namespace of the application set in the ArgoCD server
	AppsetNamespace *string `json:"appsetNamespace,omitempty"`
}

type ApplicationSetIgnoreDifferences []ApplicationSetResourceIgnoreDifferences

type ApplicationSetResourceIgnoreDifferences struct {
	// Name is the name of the application to ignore differences for. If not specified, the rule applies to all applications.
	Name string `json:"name,omitempty" protobuf:"bytes,1,name=name"`
	// JSONPointers is a list of JSON pointers to fields to ignore differences for.
	JSONPointers []string `json:"jsonPointers,omitempty" protobuf:"bytes,2,name=jsonPointers"`
	// JQPathExpressions is a list of JQ path expressions to fields to ignore differences for.
	JQPathExpressions []string `json:"jqPathExpressions,omitempty" protobuf:"bytes,3,name=jqExpressions"`
}

// ApplicationPreservedFields ApplicationSetObservation are the preseverable fields on an Application
type ApplicationPreservedFields struct {
	Annotations []string `json:"annotations,omitempty" protobuf:"bytes,1,name=annotations"`
	Labels      []string `json:"labels,omitempty" protobuf:"bytes,2,name=labels"`
}

// ApplicationSetStrategy configures how generated Applications are updated in sequence.
type ApplicationSetStrategy struct {
	Type        string                         `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`
	RollingSync *ApplicationSetRolloutStrategy `json:"rollingSync,omitempty" protobuf:"bytes,2,opt,name=rollingSync"`
	// RollingUpdate *ApplicationSetRolloutStrategy `json:"rollingUpdate,omitempty" protobuf:"bytes,3,opt,name=rollingUpdate"`
}

// ApplicationSetRolloutStrategy define the rollout strategy for the ApplicationSet
type ApplicationSetRolloutStrategy struct {
	Steps []ApplicationSetRolloutStep `json:"steps,omitempty" protobuf:"bytes,1,opt,name=steps"`
}

// ApplicationSetRolloutStep define the rollout step for the ApplicationSet
type ApplicationSetRolloutStep struct {
	MatchExpressions []ApplicationMatchExpression `json:"matchExpressions,omitempty" protobuf:"bytes,1,opt,name=matchExpressions"`
	MaxUpdate        *intstr.IntOrString          `json:"maxUpdate,omitempty" protobuf:"bytes,2,opt,name=maxUpdate"`
}

// ApplicationMatchExpression define expressions to match Applications
type ApplicationMatchExpression struct {
	Key      string   `json:"key,omitempty" protobuf:"bytes,1,opt,name=key"`
	Operator string   `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator"`
	Values   []string `json:"values,omitempty" protobuf:"bytes,3,opt,name=values"`
}

// ApplicationSetGenerator defines the generators for the ApplicationSet
type ApplicationSetGenerator struct {
	List                    *ListGenerator        `json:"list,omitempty" protobuf:"bytes,1,name=list"`
	Clusters                *ClusterGenerator     `json:"clusters,omitempty" protobuf:"bytes,2,name=clusters"`
	Git                     *GitGenerator         `json:"git,omitempty" protobuf:"bytes,3,name=git"`
	SCMProvider             *SCMProviderGenerator `json:"scmProvider,omitempty" protobuf:"bytes,4,name=scmProvider"`
	ClusterDecisionResource *DuckTypeGenerator    `json:"clusterDecisionResource,omitempty" protobuf:"bytes,5,name=clusterDecisionResource"`
	PullRequest             *PullRequestGenerator `json:"pullRequest,omitempty" protobuf:"bytes,6,name=pullRequest"`
	Matrix                  *MatrixGenerator      `json:"matrix,omitempty" protobuf:"bytes,7,name=matrix"`
	Merge                   *MergeGenerator       `json:"merge,omitempty" protobuf:"bytes,8,name=merge"`

	// Selector allows to post-filter all generator.
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,9,name=selector"`

	Plugin *PluginGenerator `json:"plugin,omitempty" protobuf:"bytes,10,name=plugin"`
}

// ApplicationSetSyncPolicy configures how generated Applications will relate to their
// ApplicationSet.
type ApplicationSetSyncPolicy struct {
	// PreserveResourcesOnDeletion will preserve resources on deletion. If PreserveResourcesOnDeletion is set to true, these Applications will not be deleted.
	PreserveResourcesOnDeletion bool `json:"preserveResourcesOnDeletion,omitempty" protobuf:"bytes,1,name=syncPolicy"`
	// ApplicationsSync represents the policy applied on the generated applications. Possible values are create-only, create-update, create-delete, sync
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=create-only;create-update;create-delete;sync
	ApplicationsSync *ApplicationsSyncPolicy `json:"applicationsSync,omitempty" protobuf:"bytes,2,opt,name=applicationsSync,casttype=ApplicationsSyncPolicy"`
}

// ApplicationsSyncPolicy representation
// "create-only" means applications are only created. If the generator's result contains update, applications won't be updated
// "create-update" means applications are only created/Updated. If the generator's result contains update, applications will be updated, but not deleted
// "create-delete" means applications are only created/deleted. If the generator's result contains update, applications won't be updated, if it results in deleted applications, the applications will be deleted
// "sync" means create/update/deleted. If the generator's result contains update, applications will be updated, if it results in deleted applications, the applications will be deleted
// If no ApplicationsSyncPolicy is defined, it defaults it to sync
type ApplicationsSyncPolicy string

// MergeGenerator merges the output of two or more generators. Where the values for all specified merge keys are equal
// between two sets of generated parameters, the parameter sets will be merged with the parameters from the latter
// generator taking precedence. Parameter sets with merge keys not present in the base generator's params will be
// ignored.
// For example, if the first generator produced [{a: '1', b: '2'}, {c: '1', d: '1'}] and the second generator produced
// [{'a': 'override'}], the united parameters for merge keys = ['a'] would be
// [{a: 'override', b: '1'}, {c: '1', d: '1'}].
//
// MergeGenerator supports template overriding. If a MergeGenerator is one of multiple top-level generators, its
// template will be merged with the top-level generator before the parameters are applied.
type MergeGenerator struct {
	Generators []ApplicationSetNestedGenerator `json:"generators" protobuf:"bytes,1,name=generators"`
	MergeKeys  []string                        `json:"mergeKeys" protobuf:"bytes,2,name=mergeKeys"`
	Template   ApplicationSetTemplate          `json:"template,omitempty" protobuf:"bytes,3,name=template"`
}

// ListGenerator include items info
type ListGenerator struct {
	Elements     []apiextv1.JSON        `json:"elements" protobuf:"bytes,1,name=elements"`
	Template     ApplicationSetTemplate `json:"template,omitempty" protobuf:"bytes,2,name=template"`
	ElementsYaml string                 `json:"elementsYaml,omitempty" protobuf:"bytes,3,opt,name=elementsYaml"`
}

// MatrixGenerator generates the cartesian product of two sets of parameters. The parameters are defined by two nested
// generators.
type MatrixGenerator struct {
	Generators []ApplicationSetNestedGenerator `json:"generators" protobuf:"bytes,1,name=generators"`
	Template   ApplicationSetTemplate          `json:"template,omitempty" protobuf:"bytes,2,name=template"`
}

// ApplicationSetNestedGenerator represents a generator nested within a combination-type generator (MatrixGenerator or
// MergeGenerator).
type ApplicationSetNestedGenerator struct {
	List                    *ListGenerator        `json:"list,omitempty" protobuf:"bytes,1,name=list"`
	Clusters                *ClusterGenerator     `json:"clusters,omitempty" protobuf:"bytes,2,name=clusters"`
	Git                     *GitGenerator         `json:"git,omitempty" protobuf:"bytes,3,name=git"`
	SCMProvider             *SCMProviderGenerator `json:"scmProvider,omitempty" protobuf:"bytes,4,name=scmProvider"`
	ClusterDecisionResource *DuckTypeGenerator    `json:"clusterDecisionResource,omitempty" protobuf:"bytes,5,name=clusterDecisionResource"`
	PullRequest             *PullRequestGenerator `json:"pullRequest,omitempty" protobuf:"bytes,6,name=pullRequest"`

	// Matrix should have the form of NestedMatrixGenerator
	Matrix *apiextv1.JSON `json:"matrix,omitempty" protobuf:"bytes,7,name=matrix"`

	// Merge should have the form of NestedMergeGenerator
	Merge *apiextv1.JSON `json:"merge,omitempty" protobuf:"bytes,8,name=merge"`

	// Selector allows to post-filter all generator.
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,9,name=selector"`

	Plugin *PluginGenerator `json:"plugin,omitempty" protobuf:"bytes,10,name=plugin"`
}

// ApplicationSetNestedGenerators represents a generator nested within a combination-type generator
type ApplicationSetNestedGenerators []ApplicationSetNestedGenerator

// NestedMatrixGenerator is a MatrixGenerator nested under another combination-type generator (MatrixGenerator or
// MergeGenerator). NestedMatrixGenerator does not have an override template, because template overriding has no meaning
// within the constituent generators of combination-type generators.
//
// NOTE: Nested matrix generator is not included directly in the CRD struct, instead it is included
// as a generic 'apiextv1.JSON' object, and then marshalled into a NestedMatrixGenerator
// when processed.
type NestedMatrixGenerator struct {
	Generators ApplicationSetTerminalGenerators `json:"generators" protobuf:"bytes,1,name=generators"`
}

// ApplicationSetTerminalGenerators represents a generator nested within a nested generator
type ApplicationSetTerminalGenerators []ApplicationSetTerminalGenerator

// ApplicationSetTerminalGenerator represents a generator nested within a nested generator (for example, a list within
// a merge within a matrix). A generator at this level may not be a combination-type generator (MatrixGenerator or
// MergeGenerator). ApplicationSet enforces this nesting depth limit because CRDs do not support recursive types.
// https://github.com/kubernetes-sigs/controller-tools/issues/477
type ApplicationSetTerminalGenerator struct {
	List                    *ListGenerator        `json:"list,omitempty" protobuf:"bytes,1,name=list"`
	Clusters                *ClusterGenerator     `json:"clusters,omitempty" protobuf:"bytes,2,name=clusters"`
	Git                     *GitGenerator         `json:"git,omitempty" protobuf:"bytes,3,name=git"`
	SCMProvider             *SCMProviderGenerator `json:"scmProvider,omitempty" protobuf:"bytes,4,name=scmProvider"`
	ClusterDecisionResource *DuckTypeGenerator    `json:"clusterDecisionResource,omitempty" protobuf:"bytes,5,name=clusterDecisionResource"`
	PullRequest             *PullRequestGenerator `json:"pullRequest,omitempty" protobuf:"bytes,6,name=pullRequest"`
	Plugin                  *PluginGenerator      `json:"plugin,omitempty" protobuf:"bytes,7,name=pullRequest"`

	// Selector allows to post-filter all generator.
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,8,name=selector"`
}

// ClusterGenerator defines a generator to match against clusters registered with ArgoCD.
type ClusterGenerator struct {
	// Selector defines a label selector to match against all clusters registered with ArgoCD.
	// Clusters today are stored as Kubernetes Secrets, thus the Secret labels will be used
	// for matching the selector.
	Selector metav1.LabelSelector   `json:"selector,omitempty" protobuf:"bytes,1,name=selector"`
	Template ApplicationSetTemplate `json:"template,omitempty" protobuf:"bytes,2,name=template"`

	// Values contains key/value pairs which are passed directly as parameters to the template
	Values map[string]string `json:"values,omitempty" protobuf:"bytes,3,name=values"`

	// returns the clusters a single 'clusters' value in the template
	FlatList bool `json:"flatList,omitempty" protobuf:"bytes,4,name=flatList"`
}

// DuckTypeGenerator defines a generator to match against clusters registered with ArgoCD.
type DuckTypeGenerator struct {
	// ConfigMapRef is a ConfigMap with the duck type definitions needed to retrieve the data
	//              this includes apiVersion(group/version), kind, matchKey and validation settings
	// Name is the resource name of the kind, group and version, defined in the ConfigMapRef
	// RequeueAfterSeconds is how long before the duckType will be rechecked for a change
	ConfigMapRef        string               `json:"configMapRef" protobuf:"bytes,1,name=configMapRef"`
	Name                string               `json:"name,omitempty" protobuf:"bytes,2,name=name"`
	RequeueAfterSeconds *int64               `json:"requeueAfterSeconds,omitempty" protobuf:"bytes,3,name=requeueAfterSeconds"`
	LabelSelector       metav1.LabelSelector `json:"labelSelector,omitempty" protobuf:"bytes,4,name=labelSelector"`

	Template ApplicationSetTemplate `json:"template,omitempty" protobuf:"bytes,5,name=template"`
	// Values contains key/value pairs which are passed directly as parameters to the template
	Values map[string]string `json:"values,omitempty" protobuf:"bytes,6,name=values"`
}

// GitGenerator defines a generator that scrapes a Git repo to find candidate resources.
type GitGenerator struct {
	RepoURL             string                      `json:"repoURL" protobuf:"bytes,1,name=repoURL"`
	Directories         []GitDirectoryGeneratorItem `json:"directories,omitempty" protobuf:"bytes,2,name=directories"`
	Files               []GitFileGeneratorItem      `json:"files,omitempty" protobuf:"bytes,3,name=files"`
	Revision            string                      `json:"revision" protobuf:"bytes,4,name=revision"`
	RequeueAfterSeconds *int64                      `json:"requeueAfterSeconds,omitempty" protobuf:"bytes,5,name=requeueAfterSeconds"`
	Template            ApplicationSetTemplate      `json:"template,omitempty" protobuf:"bytes,6,name=template"`
	PathParamPrefix     string                      `json:"pathParamPrefix,omitempty" protobuf:"bytes,7,name=pathParamPrefix"`

	// Values contains key/value pairs which are passed directly as parameters to the template
	Values map[string]string `json:"values,omitempty" protobuf:"bytes,8,name=values"`
}

// GitDirectoryGeneratorItem defines a directory to scan for resources.
type GitDirectoryGeneratorItem struct {
	Path    string `json:"path" protobuf:"bytes,1,name=path"`
	Exclude bool   `json:"exclude,omitempty" protobuf:"bytes,2,name=exclude"`
}

// GitFileGeneratorItem defines a file to scan for resources.
type GitFileGeneratorItem struct {
	Path string `json:"path" protobuf:"bytes,1,name=path"`
}

// SCMProviderGenerator defines a generator that scrapes a SCMaaS API to find candidate repos.
type SCMProviderGenerator struct {
	// Which provider to use and config for it.
	Github          *SCMProviderGeneratorGithub          `json:"github,omitempty" protobuf:"bytes,1,opt,name=github"`
	Gitlab          *SCMProviderGeneratorGitlab          `json:"gitlab,omitempty" protobuf:"bytes,2,opt,name=gitlab"`
	Bitbucket       *SCMProviderGeneratorBitbucket       `json:"bitbucket,omitempty" protobuf:"bytes,3,opt,name=bitbucket"`
	BitbucketServer *SCMProviderGeneratorBitbucketServer `json:"bitbucketServer,omitempty" protobuf:"bytes,4,opt,name=bitbucketServer"`
	Gitea           *SCMProviderGeneratorGitea           `json:"gitea,omitempty" protobuf:"bytes,5,opt,name=gitea"`
	AzureDevOps     *SCMProviderGeneratorAzureDevOps     `json:"azureDevOps,omitempty" protobuf:"bytes,6,opt,name=azureDevOps"`
	// Filters for which repos should be considered.
	Filters []SCMProviderGeneratorFilter `json:"filters,omitempty" protobuf:"bytes,7,rep,name=filters"`
	// Which protocol to use for the SCM URL. Default is provider-specific but ssh if possible. Not all providers
	// necessarily support all protocols.
	CloneProtocol string `json:"cloneProtocol,omitempty" protobuf:"bytes,8,opt,name=cloneProtocol"`
	// Standard parameters.
	RequeueAfterSeconds *int64                 `json:"requeueAfterSeconds,omitempty" protobuf:"varint,9,opt,name=requeueAfterSeconds"`
	Template            ApplicationSetTemplate `json:"template,omitempty" protobuf:"bytes,10,opt,name=template"`

	// Values contains key/value pairs which are passed directly as parameters to the template
	Values        map[string]string                  `json:"values,omitempty" protobuf:"bytes,11,name=values"`
	AWSCodeCommit *SCMProviderGeneratorAWSCodeCommit `json:"awsCodeCommit,omitempty" protobuf:"bytes,12,opt,name=awsCodeCommit"`
}

// SCMProviderGeneratorGitea defines a connection info specific to Gitea.
type SCMProviderGeneratorGitea struct {
	// Gitea organization or user to scan. Required.
	Owner string `json:"owner" protobuf:"bytes,1,opt,name=owner"`
	// The Gitea URL to talk to. For example https://gitea.mydomain.com/.
	API string `json:"api" protobuf:"bytes,2,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,3,opt,name=tokenRef"`
	// Scan all branches instead of just the default branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,4,opt,name=allBranches"`
	// Allow self-signed TLS / Certificates; default: false
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,5,opt,name=insecure"`
}

// SCMProviderGeneratorGithub defines connection info specific to GitHub.
type SCMProviderGeneratorGithub struct {
	// GitHub org to scan. Required.
	Organization string `json:"organization" protobuf:"bytes,1,opt,name=organization"`
	// The GitHub API URL to talk to. If blank, use https://api.github.com/.
	API string `json:"api,omitempty" protobuf:"bytes,2,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,3,opt,name=tokenRef"`
	// AppSecretName is a reference to a GitHub App repo-creds secret.
	AppSecretName string `json:"appSecretName,omitempty" protobuf:"bytes,4,opt,name=appSecretName"`
	// Scan all branches instead of just the default branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,5,opt,name=allBranches"`
}

// SCMProviderGeneratorGitlab defines connection info specific to Gitlab.
type SCMProviderGeneratorGitlab struct {
	// Gitlab group to scan. Required.  You can use either the project id (recommended) or the full namespaced path.
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Recurse through subgroups (true) or scan only the base group (false).  Defaults to "false"
	IncludeSubgroups bool `json:"includeSubgroups,omitempty" protobuf:"varint,2,opt,name=includeSubgroups"`
	// The Gitlab API URL to talk to.
	API string `json:"api,omitempty" protobuf:"bytes,3,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,4,opt,name=tokenRef"`
	// Scan all branches instead of just the default branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,5,opt,name=allBranches"`
	// Skips validating the SCM provider's TLS certificate - useful for self-signed certificates.; default: false
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,6,opt,name=insecure"`
	// When recursing through subgroups, also include shared Projects (true) or scan only the subgroups under same path (false).  Defaults to "true"
	IncludeSharedProjects *bool `json:"includeSharedProjects,omitempty" protobuf:"varint,7,opt,name=includeSharedProjects"`
	// Filter repos list based on Gitlab Topic.
	Topic *string `json:"topic,omitempty" protobuf:"bytes,8,opt,name=topic"`
	// ConfigMap key holding the trusted certificates
	CARef *ConfigMapKeyRef `json:"caRef,omitempty" protobuf:"bytes,9,opt,name=caRef"`
}

// SCMProviderGeneratorBitbucket defines connection info specific to Bitbucket Cloud (API version 2).
type SCMProviderGeneratorBitbucket struct {
	// Bitbucket workspace to scan. Required.
	Owner string `json:"owner" protobuf:"bytes,1,opt,name=owner"`
	// Bitbucket user to use when authenticating.  Should have a "member" role to be able to read all repositories and branches.  Required
	User string `json:"user" protobuf:"bytes,2,opt,name=user"`
	// The app password to use for the user.  Required. See: https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/
	AppPasswordRef *SecretRef `json:"appPasswordRef" protobuf:"bytes,3,opt,name=appPasswordRef"`
	// Scan all branches instead of just the main branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,4,opt,name=allBranches"`
}

// SCMProviderGeneratorBitbucketServer defines connection info specific to Bitbucket Server.
type SCMProviderGeneratorBitbucketServer struct {
	// Project to scan. Required.
	Project string `json:"project" protobuf:"bytes,1,opt,name=project"`
	// The Bitbucket Server REST API URL to talk to. Required.
	API string `json:"api" protobuf:"bytes,2,opt,name=api"`
	// Credentials for Basic auth
	BasicAuth *BasicAuthBitbucketServer `json:"basicAuth,omitempty" protobuf:"bytes,3,opt,name=basicAuth"`
	// Scan all branches instead of just the default branch.
	AllBranches *bool `json:"allBranches,omitempty" protobuf:"varint,4,opt,name=allBranches"`
	// Credentials for AccessToken (Bearer auth)
	BearerToken *BearerTokenBitbucket `json:"bearerToken,omitempty" protobuf:"bytes,5,opt,name=bearerToken"`
	// Allow self-signed TLS / Certificates; default: false
	Insecure *bool `json:"insecure,omitempty" protobuf:"varint,6,opt,name=insecure"`
	// ConfigMap key holding the trusted certificates
	CARef *ConfigMapKeyRef `json:"caRef,omitempty" protobuf:"bytes,7,opt,name=caRef"`
}

type BearerTokenBitbucket struct {
	// Password (or personal access token) reference.
	TokenRef *SecretRef `json:"tokenRef" protobuf:"bytes,1,opt,name=tokenRef"`
}

// SCMProviderGeneratorAzureDevOps defines connection info specific to Azure DevOps.
type SCMProviderGeneratorAzureDevOps struct {
	// Azure Devops organization. Required. E.g. "my-organization".
	Organization string `json:"organization" protobuf:"bytes,5,opt,name=organization"`
	// The URL to Azure DevOps. If blank, use https://dev.azure.com.
	API string `json:"api,omitempty" protobuf:"bytes,6,opt,name=api"`
	// Azure Devops team project. Required. E.g. "my-team".
	TeamProject string `json:"teamProject" protobuf:"bytes,7,opt,name=teamProject"`
	// The Personal Access Token (PAT) to use when connecting. Required.
	AccessTokenRef *SecretRef `json:"accessTokenRef" protobuf:"bytes,8,opt,name=accessTokenRef"`
	// Scan all branches instead of just the default branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,9,opt,name=allBranches"`
}

// TagFilter defines tags to filter from
type TagFilter struct {
	Key   string `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

// SCMProviderGeneratorAWSCodeCommit defines connection info specific to AWS CodeCommit.
type SCMProviderGeneratorAWSCodeCommit struct {
	// TagFilters provides the tag filter(s) for repo discovery
	TagFilters []*TagFilter `json:"tagFilters,omitempty" protobuf:"bytes,1,opt,name=tagFilters"`
	// Role provides the AWS IAM role to assume, for cross-account repo discovery
	// if not provided, AppSet controller will use its pod/node identity to discover.
	Role string `json:"role,omitempty" protobuf:"bytes,2,opt,name=role"`
	// Region provides the AWS region to discover repos.
	// if not provided, AppSet controller will infer the current region from environment.
	Region string `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`
	// Scan all branches instead of just the default branch.
	AllBranches bool `json:"allBranches,omitempty" protobuf:"varint,4,opt,name=allBranches"`
}

// SCMProviderGeneratorFilter is a single repository filter.
// If multiple filter types are set on a single struct, they will be AND'd together. All filters must
// pass for a repo to be included.
type SCMProviderGeneratorFilter struct {
	// A regex for repo names.
	RepositoryMatch *string `json:"repositoryMatch,omitempty" protobuf:"bytes,1,opt,name=repositoryMatch"`
	// An array of paths, all of which must exist.
	PathsExist []string `json:"pathsExist,omitempty" protobuf:"bytes,2,rep,name=pathsExist"`
	// An array of paths, all of which must not exist.
	PathsDoNotExist []string `json:"pathsDoNotExist,omitempty" protobuf:"bytes,3,rep,name=pathsDoNotExist"`
	// A regex which must match at least one label.
	LabelMatch *string `json:"labelMatch,omitempty" protobuf:"bytes,4,opt,name=labelMatch"`
	// A regex which must match the branch name.
	BranchMatch *string `json:"branchMatch,omitempty" protobuf:"bytes,5,opt,name=branchMatch"`
}

// PullRequestGenerator defines a generator that scrapes a PullRequest API to find candidate pull requests.
type PullRequestGenerator struct {
	// Which provider to use and config for it.
	Github          *PullRequestGeneratorGithub          `json:"github,omitempty" protobuf:"bytes,1,opt,name=github"`
	GitLab          *PullRequestGeneratorGitLab          `json:"gitlab,omitempty" protobuf:"bytes,2,opt,name=gitlab"`
	Gitea           *PullRequestGeneratorGitea           `json:"gitea,omitempty" protobuf:"bytes,3,opt,name=gitea"`
	BitbucketServer *PullRequestGeneratorBitbucketServer `json:"bitbucketServer,omitempty" protobuf:"bytes,4,opt,name=bitbucketServer"`
	// Filters for which pull requests should be considered.
	Filters []PullRequestGeneratorFilter `json:"filters,omitempty" protobuf:"bytes,5,rep,name=filters"`
	// Standard parameters.
	RequeueAfterSeconds *int64                         `json:"requeueAfterSeconds,omitempty" protobuf:"varint,6,opt,name=requeueAfterSeconds"`
	Template            ApplicationSetTemplate         `json:"template,omitempty" protobuf:"bytes,7,opt,name=template"`
	Bitbucket           *PullRequestGeneratorBitbucket `json:"bitbucket,omitempty" protobuf:"bytes,8,opt,name=bitbucket"`
	// Additional provider to use and config for it.
	AzureDevOps *PullRequestGeneratorAzureDevOps `json:"azuredevops,omitempty" protobuf:"bytes,9,opt,name=azuredevops"`

	// Values contains key/value pairs which are passed directly as parameters to the template
	Values map[string]string `json:"values,omitempty" protobuf:"bytes,10,name=values"`

	// If you add a new SCM provider, update CustomApiUrl below.
}

// PullRequestGeneratorGitea defines connection info specific to Gitea.
type PullRequestGeneratorGitea struct {
	// Gitea org or user to scan. Required.
	Owner string `json:"owner" protobuf:"bytes,1,opt,name=owner"`
	// Gitea repo name to scan. Required.
	Repo string `json:"repo" protobuf:"bytes,2,opt,name=repo"`
	// The Gitea API URL to talk to. Required
	API string `json:"api" protobuf:"bytes,3,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,4,opt,name=tokenRef"`
	// Allow insecure tls, for self-signed certificates; default: false.
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,5,opt,name=insecure"`
}

// PullRequestGeneratorAzureDevOps defines connection info specific to AzureDevOps.
type PullRequestGeneratorAzureDevOps struct {
	// Azure DevOps org to scan. Required.
	Organization string `json:"organization" protobuf:"bytes,1,opt,name=organization"`
	// Azure DevOps project name to scan. Required.
	Project string `json:"project" protobuf:"bytes,2,opt,name=project"`
	// Azure DevOps repo name to scan. Required.
	Repo string `json:"repo" protobuf:"bytes,3,opt,name=repo"`
	// The Azure DevOps API URL to talk to. If blank, use https://dev.azure.com/.
	API string `json:"api,omitempty" protobuf:"bytes,4,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,5,opt,name=tokenRef"`
	// Labels is used to filter the PRs that you want to target
	Labels []string `json:"labels,omitempty" protobuf:"bytes,6,rep,name=labels"`
}

// PullRequestGeneratorGithub defines connection info specific to GitHub.
type PullRequestGeneratorGithub struct {
	// GitHub org or user to scan. Required.
	Owner string `json:"owner" protobuf:"bytes,1,opt,name=owner"`
	// GitHub repo name to scan. Required.
	Repo string `json:"repo" protobuf:"bytes,2,opt,name=repo"`
	// The GitHub API URL to talk to. If blank, use https://api.github.com/.
	API string `json:"api,omitempty" protobuf:"bytes,3,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,4,opt,name=tokenRef"`
	// AppSecretName is a reference to a GitHub App repo-creds secret with permission to access pull requests.
	AppSecretName string `json:"appSecretName,omitempty" protobuf:"bytes,5,opt,name=appSecretName"`
	// Labels is used to filter the PRs that you want to target
	Labels []string `json:"labels,omitempty" protobuf:"bytes,6,rep,name=labels"`
}

// PullRequestGeneratorGitLab defines connection info specific to GitLab.
type PullRequestGeneratorGitLab struct {
	// GitLab project to scan. Required.
	Project string `json:"project" protobuf:"bytes,1,opt,name=project"`
	// The GitLab API URL to talk to. If blank, uses https://gitlab.com/.
	API *string `json:"api,omitempty" protobuf:"bytes,2,opt,name=api"`
	// Authentication token reference.
	TokenRef *SecretRef `json:"tokenRef,omitempty" protobuf:"bytes,3,opt,name=tokenRef"`
	// Labels is used to filter the MRs that you want to target
	Labels []string `json:"labels,omitempty" protobuf:"bytes,4,rep,name=labels"`
	// PullRequestState is an additional MRs filter to get only those with a certain state. Default: "" (all states)
	PullRequestState *string `json:"pullRequestState,omitempty" protobuf:"bytes,5,rep,name=pullRequestState"`
	// Skips validating the SCM provider's TLS certificate - useful for self-signed certificates.; default: false
	Insecure *bool `json:"insecure,omitempty" protobuf:"varint,6,opt,name=insecure"`
	// ConfigMap key holding the trusted certificates
	CARef *ConfigMapKeyRef `json:"caRef,omitempty" protobuf:"bytes,7,opt,name=caRef"`
}

// PullRequestGeneratorBitbucketServer defines connection info specific to BitbucketServer.
type PullRequestGeneratorBitbucketServer struct {
	// Project to scan. Required.
	Project string `json:"project" protobuf:"bytes,1,opt,name=project"`
	// Repo name to scan. Required.
	Repo string `json:"repo" protobuf:"bytes,2,opt,name=repo"`
	// The Bitbucket REST API URL to talk to e.g. https://bitbucket.org/rest Required.
	API string `json:"api" protobuf:"bytes,3,opt,name=api"`
	// Credentials for Basic auth
	BasicAuth *BasicAuthBitbucketServer `json:"basicAuth,omitempty" protobuf:"bytes,4,opt,name=basicAuth"`
	// Credentials for AccessToken (Bearer auth)
	BearerToken *BearerTokenBitbucket `json:"bearerToken,omitempty" protobuf:"bytes,5,opt,name=bearerToken"`
	// Allow self-signed TLS / Certificates; default: false
	Insecure *bool `json:"insecure,omitempty" protobuf:"varint,6,opt,name=insecure"`
	// ConfigMap key holding the trusted certificates
	CARef *ConfigMapKeyRef `json:"caRef,omitempty" protobuf:"bytes,7,opt,name=caRef"`
}

// PullRequestGeneratorBitbucket defines connection info specific to Bitbucket.
type PullRequestGeneratorBitbucket struct {
	// Workspace to scan. Required.
	Owner string `json:"owner" protobuf:"bytes,1,opt,name=owner"`
	// Repo name to scan. Required.
	Repo string `json:"repo" protobuf:"bytes,2,opt,name=repo"`
	// The Bitbucket REST API URL to talk to. If blank, uses https://api.bitbucket.org/2.0.
	API string `json:"api,omitempty" protobuf:"bytes,3,opt,name=api"`
	// Credentials for Basic auth
	BasicAuth *BasicAuthBitbucketServer `json:"basicAuth,omitempty" protobuf:"bytes,4,opt,name=basicAuth"`
	// Credentials for AppToken (Bearer auth)
	BearerToken *BearerTokenBitbucketCloud `json:"bearerToken,omitempty" protobuf:"bytes,5,opt,name=bearerToken"`
}

// BearerTokenBitbucketCloud defines the Bearer token for BitBucket AppToken auth.
type BearerTokenBitbucketCloud struct {
	// Password (or personal access token) reference.
	TokenRef *SecretRef `json:"tokenRef" protobuf:"bytes,1,opt,name=tokenRef"`
}

// BasicAuthBitbucketServer defines the username/(password or personal access token) for Basic auth.
type BasicAuthBitbucketServer struct {
	// Username for Basic auth
	Username string `json:"username" protobuf:"bytes,1,opt,name=username"`
	// Password (or personal access token) reference.
	PasswordRef *SecretRef `json:"passwordRef" protobuf:"bytes,2,opt,name=passwordRef"`
}

// PullRequestGeneratorFilter is a single pull request filter.
// If multiple filter types are set on a single struct, they will be AND'd together. All filters must
// pass for a pull request to be included.
type PullRequestGeneratorFilter struct {
	BranchMatch       *string `json:"branchMatch,omitempty" protobuf:"bytes,1,opt,name=branchMatch"`
	TargetBranchMatch *string `json:"targetBranchMatch,omitempty" protobuf:"bytes,2,opt,name=targetBranchMatch"`
}

// SecretRef is a utility struct for a reference to a secret key.
type SecretRef struct {
	SecretName string `json:"secretName" protobuf:"bytes,1,opt,name=secretName"`
	Key        string `json:"key" protobuf:"bytes,2,opt,name=key"`
}

// ConfigMapKeyRef is a utility struct for a reference to a configmap key.
type ConfigMapKeyRef struct {
	ConfigMapName string `json:"configMapName" protobuf:"bytes,1,opt,name=configMapName"`
	Key           string `json:"key" protobuf:"bytes,2,opt,name=key"`
}

// PluginConfigMapRef defines a reference to a ConfigMap containing a plugin.
type PluginConfigMapRef struct {
	// Name of the ConfigMap
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

// PluginParameters defines the parameters to pass to a plugin.
type PluginParameters map[string]apiextv1.JSON

// PluginInput defines the input to a plugin.
type PluginInput struct {
	// Parameters contains the information to pass to the plugin. It is a map. The keys must be strings, and the
	// values can be any type.
	Parameters PluginParameters `json:"parameters,omitempty" protobuf:"bytes,1,name=parameters"`
}

// PluginGenerator defines connection info specific to Plugin.
type PluginGenerator struct {
	ConfigMapRef PluginConfigMapRef `json:"configMapRef" protobuf:"bytes,1,name=configMapRef"`
	Input        PluginInput        `json:"input,omitempty" protobuf:"bytes,2,name=input"`
	// RequeueAfterSeconds determines how long the ApplicationSet controller will wait before reconciling the ApplicationSet again.
	RequeueAfterSeconds *int64                 `json:"requeueAfterSeconds,omitempty" protobuf:"varint,3,opt,name=requeueAfterSeconds"`
	Template            ApplicationSetTemplate `json:"template,omitempty" protobuf:"bytes,4,name=template"`

	// Values contains key/value pairs which are passed directly as parameters to the template. These values will not be
	// sent as parameters to the plugin.
	Values map[string]string `json:"values,omitempty" protobuf:"bytes,5,name=values"`
}

// ApplicationSetTemplate represents argocd ApplicationSpec
type ApplicationSetTemplate struct {
	ApplicationSetTemplateMeta `json:"metadata" protobuf:"bytes,1,name=metadata"`
	Spec                       ApplicationSpec `json:"spec" protobuf:"bytes,2,name=spec"`
}

// ApplicationSetTemplateMeta represents the Argo CD application fields that may
// be used for Applications generated from the ApplicationSet (based on metav1.ObjectMeta)
type ApplicationSetTemplateMeta struct {
	Name        string            `json:"name,omitempty" protobuf:"bytes,1,name=name"`
	Namespace   string            `json:"namespace,omitempty" protobuf:"bytes,2,name=namespace"`
	Labels      map[string]string `json:"labels,omitempty" protobuf:"bytes,3,name=labels"`
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,4,name=annotations"`
	Finalizers  []string          `json:"finalizers,omitempty" protobuf:"bytes,5,name=finalizers"`
}

// A ApplicationSetSpec defines the desired state of a ApplicationSet.
type ApplicationSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationSetParameters `json:"forProvider"`
}

// A ApplicationSetStatus represents the observed state of a ApplicationSet.
type ApplicationSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ArgoApplicationSetStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ApplicationSet is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,argocd}
type ApplicationSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSetSpec   `json:"spec"`
	Status ApplicationSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationSetList contains a list of ApplicationSet
type ApplicationSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationSet `json:"items"`
}
