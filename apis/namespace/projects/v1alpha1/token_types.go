package v1alpha1

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clusterapis "github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
)

// A TokenSpec defines the desired state of an ArgoCD Token.
type TokenSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       clusterapis.TokenParameters `json:"forProvider"`
}

// A TokenStatus represents the observed state of an ArgoCD Project Token.
type TokenStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          clusterapis.TokenObservation `json:"atProvider,omitempty"`
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
// +kubebuilder:resource:scope=Namespace,categories={crossplane,managed,argocd}
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

// Token type metadata
var (
	TokenKind             = reflect.TypeOf(Token{}).Name()
	TokenGroupKind        = schema.GroupKind{Group: Group, Kind: TokenKind}.String()
	TokenKindAPIVersion   = TokenKind + "." + SchemeGroupVersion.String()
	TokenGroupVersionKind = SchemeGroupVersion.WithKind(TokenKind)
)

func init() {
	SchemeBuilder.Register(&Token{}, &TokenList{})
}
