package applicationsets

import (
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argoGrpc "github.com/argoproj/argo-cd/v2/util/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// ServiceClient wraps the functions to connect to argocd repositories
type ServiceClient interface {
	// Get returns an applicationset by name
	Get(ctx context.Context, in *applicationset.ApplicationSetGetQuery, opts ...grpc.CallOption) (*v1alpha1.ApplicationSet, error)
	// List returns list of applicationset
	List(ctx context.Context, in *applicationset.ApplicationSetListQuery, opts ...grpc.CallOption) (*v1alpha1.ApplicationSetList, error)
	// Create creates an applicationset
	Create(ctx context.Context, in *applicationset.ApplicationSetCreateRequest, opts ...grpc.CallOption) (*v1alpha1.ApplicationSet, error)
	// Delete deletes an application set
	Delete(ctx context.Context, in *applicationset.ApplicationSetDeleteRequest, opts ...grpc.CallOption) (*applicationset.ApplicationSetResponse, error)
}

// NewApplicationSetServiceClient creates a new API client from a set of config options, or fails fatally if the new client creation fails.
func NewApplicationSetServiceClient(clientOpts *apiclient.ClientOptions) ServiceClient {
	_, repoIf := apiclient.NewClientOrDie(clientOpts).NewApplicationSetClientOrDie()
	return repoIf
}

// IsNotFound returns true if the error code is NotFound
func IsNotFound(err error) bool {
	unwrappedError := argoGrpc.UnwrapGRPCStatus(err).Code()
	return unwrappedError == codes.NotFound
}
