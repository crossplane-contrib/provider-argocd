# provider-argocd

## Overview

`provider-argocd` is the Crossplane infrastructure provider for
[argocd](https://argo-cd.readthedocs.io/). The provider that is built from the source code
in this repository can be installed into a Crossplane control plane and adds the
following new functionality:

* Custom Resource Definitions (CRDs) that model argocd resources
* Controllers to provision these resources in ArgoCD based on the users desired
  state captured in CRDs they create
* Implementations of Crossplane's portable resource
  abstractions, enabling
  argocd resources to fulfill a user's general need for argocd configurations

## Getting Started and Documentation

Follow these steps to get started with `provider-argocd`.

### Optional: Start a Local ArgoCD Server

    kind create cluster

    kubectl create ns argocd

    kubectl apply -n argocd --force -f https://raw.githubusercontent.com/argoproj/argo-cd/release-2.0/manifests/install.yaml
    
    helm upgrade --install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace --wait


### Optional: Create a new user

Follow the steps in the [official documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/) to create a new user `provider-argcod`.

> _**To test** you can create one user by patching your argocd-configmap and argocd-rbac definition file_
> 
> _**Be careful:** this command allows all users, to do everything. So never do this in production_

    kubectl patch configmap/argocd-rbac-cm -n argocd --type merge -p '{"data":{"policy.default":"role:admin"}}'
    kubectl patch configmap/argocd-cm -n argocd --type merge -p '{"data":{"accounts.provider-argocd":"apiKey, login"}}'

### Create an API Token

*Note:* The following steps require the [kubectl-view-secret](https://github.com/elsesiy/kubectl-view-secret) plugin and [jq](https://stedolan.github.io/jq/) to be installed.

Get the admin passwort via `kubectl`
```bash
ARGOCD_ADMIN_SECRET=$(kubectl view-secret argocd-initial-admin-secret -n argocd -q)
```

Port forward the argocd api to the host:
```bash
kubectl -n argocd port-forward svc/argocd-server 8443:443
```

Create a session JWT for the admin user at the ArgoCD API. *Note:* You cannot use this token directly, because it will expire.
```bash
ARGOCD_ADMIN_TOKEN=$(curl -s -X POST -k -H "Content-Type: application/json" --data '{"username":"admin","password":"'$ARGOCD_ADMIN_SECRET'"}' https://localhost:8443/api/v1/session | jq -r .token)
```

Create an API token without expiration that can be used by `provider-argocd`
```bash
ARGOCD_PROVIDER_USER="provider-argocd"

ARGOCD_TOKEN=$(curl -s -X POST -k -H "Authorization: Bearer $ARGOCD_ADMIN_TOKEN" -H "Content-Type: application/json" https://localhost:8443/api/v1/account/$ARGOCD_PROVIDER_USER/token | jq -r .token)
```

### Setup crossplane provider-argocd

Create a kubernetes secret from the JWT so `provider-argocd` is able to connect to argocd:
```bash
kubectl create secret generic argocd-credentials -n crossplane-system --from-literal=authToken="$ARGOCD_TOKEN"
```

Install provider-argocd:

`kubectl crossplane install provider crossplane/provider-argo:v0.1.0`

Configure a `ProviderConfig` with `serverAddr` pointing to an argocd instance:
```yaml
apiVersion: argocd.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: argocd-provider
spec:
  serverAddr: argocd-server.argocd.svc:443
  insecure: true
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: argocd-credentials
      key: authToken
```
```bash
kubectl apply -f examples/providerconfig/provider.yaml
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
