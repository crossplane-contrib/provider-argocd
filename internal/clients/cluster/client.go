package cluster

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"google.golang.org/grpc"
)

const (
	errorClusterNotFound  = "code = NotFound desc = cluster"
	errorPermissionDenied = "code = PermissionDenied desc = permission denied"
)

// ServiceClient wraps the functions to connect to argocd repositories
type ServiceClient interface {
	// Create creates a cluster
	Create(ctx context.Context, in *cluster.ClusterCreateRequest, opts ...grpc.CallOption) (*v1alpha1.Cluster, error)
	// Get returns a cluster by server address
	Get(ctx context.Context, in *cluster.ClusterQuery, opts ...grpc.CallOption) (*v1alpha1.Cluster, error)
	// Update updates a cluster
	Update(ctx context.Context, in *cluster.ClusterUpdateRequest, opts ...grpc.CallOption) (*v1alpha1.Cluster, error)
	// Delete deletes a cluster
	Delete(ctx context.Context, in *cluster.ClusterQuery, opts ...grpc.CallOption) (*cluster.ClusterResponse, error)
}

// NewClusterServiceClient creates a new API client from a set of config options, or fails fatally if the new client creation fails.
func NewClusterServiceClient(clientOpts *apiclient.ClientOptions) cluster.ClusterServiceClient {
	_, repoIf := apiclient.NewClientOrDie(clientOpts).NewClusterClientOrDie()
	return repoIf
}

// IsErrorClusterNotFound helper function to test for errorClusterNotFound error.
func IsErrorClusterNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorClusterNotFound)
}

// IsErrorPermissionDenied helper function to test for errorPermissionDenied error.
func IsErrorPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorPermissionDenied)
}
