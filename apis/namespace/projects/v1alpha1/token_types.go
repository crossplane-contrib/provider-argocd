package v1alpha1

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Copy types from cluster-scope apis replace references with namespace types:
//go:generate go run -modfile ../../../../tools/go.mod -tags generate github.com/mistermx/copystruct/cmd/copystruct ../../../cluster/projects/v1alpha1 zz_generated.token_types.copied.go TokenParameters,TokenObservation
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/apis/cluster|github.com/crossplane-contrib/provider-argocd/apis/namespace|g zz_generated.token_types.copied.go
//go:generate sed -i s|v1\.Reference|v1.NamespacedReference|g zz_generated.token_types.copied.go
//go:generate sed -i s|v1\.Selector|v1.NamespacedSelector|g zz_generated.token_types.copied.go

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
