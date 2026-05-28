package applicationsets

import (
	"context"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoGrpc "github.com/argoproj/argo-cd/v3/util/grpc"
	"github.com/argoproj/argo-cd/v3/util/io"
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

// NewApplicationSetServiceClient creates a new API client from a set of config
// options. Any error from constructing the underlying argo-cd client or
// opening the application-set gRPC connection is returned to the caller so the
// reconciler can retry with backoff instead of crashing the controller process.
func NewApplicationSetServiceClient(clientOpts *apiclient.ClientOptions) (io.Closer, ServiceClient, error) {
	client, err := apiclient.NewClient(clientOpts)
	if err != nil {
		return nil, nil, err
	}
	conn, repoIf, err := client.NewApplicationSetClient()
	if err != nil {
		return nil, nil, err
	}
	return conn, repoIf, nil
}

// IsNotFound returns true if the error code is NotFound
func IsNotFound(err error) bool {
	unwrappedError := argoGrpc.UnwrapGRPCStatus(err).Code()
	return unwrappedError == codes.NotFound
}
