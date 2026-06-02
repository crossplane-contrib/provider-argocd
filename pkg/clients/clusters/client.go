package cluster

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/io"
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

// NewClusterServiceClient creates a new API client from a set of config
// options. Any error from constructing the underlying argo-cd client or
// opening the cluster gRPC connection is returned to the caller so the
// reconciler can retry with backoff instead of crashing the controller process.
func NewClusterServiceClient(clientOpts *apiclient.ClientOptions) (io.Closer, cluster.ClusterServiceClient, error) {
	client, err := apiclient.NewClient(clientOpts)
	if err != nil {
		return nil, nil, err
	}
	conn, repoIf, err := client.NewClusterClient()
	if err != nil {
		return nil, nil, err
	}
	return conn, repoIf, nil
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
