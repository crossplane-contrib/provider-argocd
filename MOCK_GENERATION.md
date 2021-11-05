# Client Mock Generation

[go-mock](https://github.com/golang/mock) is used to generate mocks of the ArgoCD client.

## Install

Follow the [installation instructions](https://github.com/golang/mock#installation) to get the latest version.

## Generate mocks

The following example shows how to generate mocks for the `projects` API:

    MOCK_API="projects"
    MOCK_INTERFACE="ProjectServiceClient"
    
    mockgen -package $MOCK_API -destination pkg/clients/mock/$MOCK_API/mock.go github.com/crossplane-contrib/provider-argocd/pkg/clients/$MOCK_API $MOCK_INTERFACE