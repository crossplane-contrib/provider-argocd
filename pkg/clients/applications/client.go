package applications

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"google.golang.org/grpc"
)

const (
	// errorApplicationNotFound = "applications.argoproj.io %s not found"
	errorApplicationNotFound = "not found"
)

type ApplicationServiceClient interface {
	// Get returns an application by name
	Get(ctx context.Context, in *application.ApplicationQuery, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// Create creates an application
	Create(ctx context.Context, in *application.ApplicationCreateRequest, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// Update updates an application
	Update(ctx context.Context, in *application.ApplicationUpdateRequest, opts ...grpc.CallOption) (*v1alpha1.Application, error)

	// Delete deletes an application
	Delete(ctx context.Context, in *application.ApplicationDeleteRequest, opts ...grpc.CallOption) (*application.ApplicationResponse, error)
}

func NewApplicationServiceClient(clientOpts *apiclient.ClientOptions) application.ApplicationServiceClient {
	_, appIf := apiclient.NewClientOrDie(clientOpts).NewApplicationClientOrDie()
	return appIf
}

// IsErrorApplicationNotFound helper function to test for errorApplicationNotFound error.
func IsErrorApplicationNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorApplicationNotFound)
}

// // IsErrorParseApplication helper function to test for errorParseApplication error.
// func IsErrorParseApplication(s string) bool {
// 	if s == "" {
// 		return false
// 	}
// 	return strings.Contains(s, errorParseApplication)
// }
