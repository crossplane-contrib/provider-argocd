package projects

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
	"google.golang.org/grpc"
)

const (
	errorProjectNotFound = "code = NotFound desc = appprojects"
)

// ProjectServiceClient wraps the functions to connect to argocd repositories
type ProjectServiceClient interface {
	// Create a new project
	Create(ctx context.Context, in *project.ProjectCreateRequest, opts ...grpc.CallOption) (*v1alpha1.AppProject, error)
	// Get returns a project by name
	Get(ctx context.Context, in *project.ProjectQuery, opts ...grpc.CallOption) (*v1alpha1.AppProject, error)
	// Update updates a project
	Update(ctx context.Context, in *project.ProjectUpdateRequest, opts ...grpc.CallOption) (*v1alpha1.AppProject, error)
	// Delete deletes a project
	Delete(ctx context.Context, in *project.ProjectQuery, opts ...grpc.CallOption) (*project.EmptyResponse, error)
	// CreateToken a new project token
	CreateToken(ctx context.Context, in *project.ProjectTokenCreateRequest, opts ...grpc.CallOption) (*project.ProjectTokenResponse, error)
	// DeleteToken a new project token
	DeleteToken(ctx context.Context, in *project.ProjectTokenDeleteRequest, opts ...grpc.CallOption) (*project.EmptyResponse, error)
}

// NewProjectServiceClient creates a new API client from a set of config
// options. Any error from constructing the underlying argo-cd client or
// opening the project gRPC connection is returned to the caller so the
// reconciler can retry with backoff instead of crashing the controller process.
func NewProjectServiceClient(clientOpts *apiclient.ClientOptions) (io.Closer, project.ProjectServiceClient, error) {
	client, err := apiclient.NewClient(clientOpts)
	if err != nil {
		return nil, nil, err
	}
	conn, repoIf, err := client.NewProjectClient()
	if err != nil {
		return nil, nil, err
	}
	return conn, repoIf, nil
}

// IsErrorProjectNotFound helper function to test for errorProjectNotFound error.
func IsErrorProjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorProjectNotFound)
}
