package tokens

//go:generate go run -modfile ../../../../tools/go.mod -tags generate github.com/mistermx/copystruct/cmd/copycode --tests ../../cluster/tokens .
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/apis/cluster|github.com/crossplane-contrib/provider-argocd/apis/namespace|g zz_generated.copied.controller.go
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/apis/cluster|github.com/crossplane-contrib/provider-argocd/apis/namespace|g zz_generated.copied.controller_test.go
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/pkg/clients/cluster|github.com/crossplane-contrib/provider-argocd/pkg/clients/namespace|g zz_generated.copied.controller.go
//go:generate sed -i s|github\.com/crossplane-contrib/provider-argocd/pkg/clients/cluster|github.com/crossplane-contrib/provider-argocd/pkg/clients/namespace|g zz_generated.copied.controller_test.go
//go:generate sed -i s|resource.ConnectionSecretFor|resource.LocalConnectionSecretFor|g zz_generated.copied.controller.go
