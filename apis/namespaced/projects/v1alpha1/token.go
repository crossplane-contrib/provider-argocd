package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TokenParameters define the desired state of an ArgoCD Project Token
type TokenParameters struct {
	// Project is the project associated with the token
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-argocd/apis/namespaced/projects/v1alpha1.Project
	// +crossplane:generate:reference:refFieldName=ProjectRef
	// +crossplane:generate:reference:selectorFieldName=ProjectSelector
	Project *string `json:"project"`

	// ProjectRefs is a reference to a Project used to set Project
	// +optional
	ProjectRef *xpv1.NamespacedReference `json:"projectRef,omitempty"`

	// ProjectSelector selects reference to a Project used to ProjectRef
	// +optional
	ProjectSelector *xpv1.NamespacedSelector `json:"projectSelector,omitempty"`

	// Role is the role associated with the token.
	Role string `json:"role"`

	// ID is an id for the token
	// +optional
	ID string `json:"id"`

	// Description is a description for the token
	// +optional
	Description *string `json:"description,omitempty"`

	// Duration before the token will expire. Valid time units are `s`, `m`, `h` and `d` E.g. 12h, 7d. No expiration if not set.
	// +optional
	// +kubebuilder:validation:Pattern=`^(0|[0-9]+(s|m|h|d))$`
	ExpiresIn *string `json:"expiresIn,omitempty"`

	// Duration to control token regeneration based on token age. Valid time units are `s`, `m`, `h` and `d`.
	// +optional
	// +kubebuilder:validation:Pattern=`^([0-9]+)(s|m|h|d)$`
	RenewAfter *string `json:"renewAfter,omitempty"`

	// Duration to control token regeneration based on remaining token lifetime. Valid time units are `s`, `m`, `h` and `d`.
	// +optional
	// +kubebuilder:validation:Pattern=`^([0-9]+)(s|m|h|d)$`
	RenewBefore *string `json:"renewBefore,omitempty"`
}

// TokenObservation holds the issuedAt and expiresAt values of a token
type TokenObservation struct {
	IssuedAt int64 `json:"iat"`
	// +optional
	ExpiresAt *int64 `json:"exp,omitempty"`
	// +optional
	ID *string `json:"id,omitempty"`
}

// A TokenSpec defines the desired state of an ArgoCD Token.
type TokenSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              TokenParameters `json:"forProvider"`
}

// A TokenStatus represents the observed state of an ArgoCD Project Token.
type TokenStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TokenObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Token is a managed resource that represents an ArgoCD Project Token
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.project"
// +kubebuilder:printcolumn:name="ROLE",type="string",JSONPath=".spec.forProvider.role"
// +kubebuilder:printcolumn:name="EXPIRES-AT",type="string",JSONPath=".status.atProvider.exp"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,argocd}
type Token struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TokenSpec   `json:"spec"`
	Status TokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TokenList contains a list of Token items
type TokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Token `json:"items"`
}
