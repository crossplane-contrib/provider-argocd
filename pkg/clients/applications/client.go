package applications

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
	"google.golang.org/grpc"
)

const (
	errorNotFound = "code = NotFound desc = repo"
)

// ServiceClient wraps the functions to connect to argocd repositories
type ServiceClient interface {
	// Get returns an application by name
	Get(ctx context.Context, in *application.ApplicationQuery, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// List returns list of applications
	List(ctx context.Context, in *application.ApplicationQuery, opts ...grpc.CallOption) (*v1alpha1.ApplicationList, error)

	// Create creates an application
	Create(ctx context.Context, in *application.ApplicationCreateRequest, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// Update updates an application
	Update(ctx context.Context, in *application.ApplicationUpdateRequest, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// Delete deletes an application
	Delete(ctx context.Context, in *application.ApplicationDeleteRequest, opts ...grpc.CallOption) (*application.ApplicationResponse, error)
}

// NewApplicationServiceClient creates a new API client from a set of config
// options. Any error from constructing the underlying argo-cd client or
// opening the application gRPC connection is returned to the caller so the
// reconciler can retry with backoff instead of crashing the controller process.
func NewApplicationServiceClient(clientOpts *apiclient.ClientOptions) (io.Closer, ServiceClient, error) {
	client, err := apiclient.NewClient(clientOpts)
	if err != nil {
		return nil, nil, err
	}
	conn, repoIf, err := client.NewApplicationClient()
	if err != nil {
		return nil, nil, err
	}
	return conn, repoIf, nil
}

// IsErrorApplicationNotFound helper function to test for errorNotFound error.
func IsErrorApplicationNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorNotFound)
}
