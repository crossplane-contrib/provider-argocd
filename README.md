# provider-argocd

## Overview

`provider-argocd` is the Crossplane infrastructure provider for
[Argo CD](https://argo-cd.readthedocs.io/). The provider that is built from the source code
in this repository can be installed into a Crossplane control plane and adds the
following new functionality:

* Custom Resource Definitions (CRDs) that model Argo CD resources
* Controllers to provision these resources in Argo CD based on the users desired
  state captured in CRDs they create
* Implementations of Crossplane's portable resource
  abstractions, enabling
  Argo CD resources to fulfill a user's general need for Argo CD configurations

## Getting Started and Documentation

Follow the [official docs](https://crossplane.io/docs/master/getting-started/install-configure.html#install-crossplane) to install crossplane, then these steps to get started with `provider-argocd`.

### Add the Crossplane Helm Repository

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
```

### Initialize Build Submodules

Before building or running the provider, ensure the required "build" Make submodule is initialized. This submodule supports CI/CD tasks shared across all providers.

```bash
make submodules
```

### Run ArgoCD and Crossplane Locally with Kind

To start a local Kubernetes cluster with `kind` and install Argo CD and Crossplane and the provider CRDs in a single command, run:

```bash
make dev-debug
```

which can later be undone with `make dev-teardown` deleting the Kind cluster.

### Run the Provider Locally for Development

To start the provider in debug mode, you can run the provider directly:

```bash
go run ./cmd/provider --debug
```

#### Optional: Run with VSCode

Alternatively, if you use VSCode, you can configure a file `.vscode/launch.json` to run the provider in debug mode in a more convenient way:

```json filename=".vscode/launch.json"
{
  "configurations": [
    {
      "name": "Run Provider Locally",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/provider",
      "args": [
        "--debug"
      ]
    }
  ]
}
```

### Apply CRs

To test the provider, you can apply the example CRs in `examples/`:

```bash
kubectl apply -f examples/projects/project.yaml
```

## Getting Started Step-by-Step

### Optional: Start a local Argo CD server
```bash
kind create cluster

kubectl create ns argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```
### Create a new Argo CD user

Follow the steps in the [official documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/) to create a new user `provider-argcod`:

```bash
kubectl patch configmap/argocd-cm \
  -n argocd \
  --type merge \
  -p '{"data":{"accounts.provider-argocd":"apiKey"}}'

kubectl patch configmap/argocd-rbac-cm \
  -n argocd \
  --type merge \
  -p '{"data":{"policy.csv":"g, provider-argocd, role:admin"}}'
```

### Create an API Token

*Note:* The following steps require the [kubectl-view-secret](https://github.com/elsesiy/kubectl-view-secret) plugin and [jq](https://stedolan.github.io/jq/) to be installed.

Get the admin passwort via `kubectl`
```bash
ARGOCD_ADMIN_SECRET=$(kubectl view-secret argocd-initial-admin-secret -n argocd -q)
```

Port forward the Argo CD api to the host:
```bash
kubectl -n argocd port-forward svc/argocd-server 8443:443
```

Create a session JWT for the admin user at the Argo CD API. *Note:* You cannot use this token directly, because it will expire.
```bash
ARGOCD_ADMIN_TOKEN=$(curl -s -X POST -k -H "Content-Type: application/json" --data '{"username":"admin","password":"'$ARGOCD_ADMIN_SECRET'"}' https://localhost:8443/api/v1/session | jq -r .token)
```

Create an API token without expiration that can be used by `provider-argocd`
```bash
ARGOCD_PROVIDER_USER="provider-argocd"

ARGOCD_TOKEN=$(curl -s -X POST -k -H "Authorization: Bearer $ARGOCD_ADMIN_TOKEN" -H "Content-Type: application/json" https://localhost:8443/api/v1/account/$ARGOCD_PROVIDER_USER/token | jq -r .token)
```

### Setup crossplane provider-argocd

Install provider-argocd:
```bash
cat << EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-argocd
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-argocd:v0.2.0
EOF
```
Create a kubernetes secret from the JWT so `provider-argocd` is able to connect to Argo CD:
```bash
kubectl create secret generic argocd-credentials -n crossplane-system --from-literal=authToken="$ARGOCD_TOKEN"
```

Configure a `ProviderConfig` with `serverAddr` pointing to an Argo CD instance:
```bash
cat << EOF | kubectl apply -f -
apiVersion: argocd.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: argocd-provider
spec:
  serverAddr: argocd-server.argocd.svc:443
  insecure: true
  plainText: false
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: argocd-credentials
      key: authToken
EOF
```

## Contributing

provider-argocd is a community driven project and we welcome contributions. See
the Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-argocd/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Governance and Owners

provider-argocd is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

provider-argocd adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-argocd is under the Apache 2.0 license.

[![FOSSA
Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-argocd.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-argocd?ref=badge_large)
