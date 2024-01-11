#!/usr/bin/env bash

# Debug
#set -x
# Be defensive
set -euo pipefail
# -m Job control is enabled
set -m

GREEN="\033[1;32m"
NOCOLOR="\033[0m"

logInfo () { echo -e  "${GREEN}[INFO] $@ ${NOCOLOR}"; }


logInfo "Waiting for Argo CD to become available."
kubectl wait --for condition=available --namespace argocd deployment.apps --all --timeout=300s

logInfo "Patching configmaps"
kubectl patch configmap/argocd-cm \
  -n argocd \
  --type merge \
  -p '{"data":{"accounts.provider-argocd":"apiKey"}}'

kubectl patch configmap/argocd-rbac-cm \
  -n argocd \
  --type merge \
  -p '{"data":{"policy.csv":"g, provider-argocd, role:admin"}}'


ARGOCD_ADMIN_SECRET=$(kubectl get secrets argocd-initial-admin-secret -n argocd  -o json | jq .data.password -r |base64 -d)
logInfo "Activating port forwarding to Argo CD to http://localhost:8443"
kubectl -n argocd port-forward svc/argocd-server 8443:443 &

logInfo "Waiting for port forwarding to become available"
while ! curl --output /dev/null --silent --head --fail http://localhost:8443; do sleep 0.1 && echo -n .; done;


ARGOCD_ADMIN_TOKEN=$(curl -s -X POST -k -H "Content-Type: application/json" --data '{"username":"admin","password":"'"$ARGOCD_ADMIN_SECRET"'"}' https://localhost:8443/api/v1/session | jq -r .token)
ARGOCD_PROVIDER_USER="provider-argocd"
ARGOCD_TOKEN=$(curl -s -X POST -k -H "Authorization: Bearer $ARGOCD_ADMIN_TOKEN" -H "Content-Type: application/json" https://localhost:8443/api/v1/account/$ARGOCD_PROVIDER_USER/token | jq -r .token)
logInfo "Creating ArgoCD provider credential"
kubectl create  secret generic --dry-run=client --save-config argocd-credentials  -n crossplane-system --from-literal=authToken="$ARGOCD_TOKEN" -o yaml | kubectl apply -f -
logInfo "Applying provider config "
cat << EOF | kubectl apply -f -
  apiVersion: argocd.crossplane.io/v1alpha1
  kind: ProviderConfig
  metadata:
    name: argocd-provider
  spec:
    serverAddr: localhost:8443
    insecure: true
    plainText: false
    credentials:
      source: Secret
      secretRef:
        namespace: crossplane-system
        name: argocd-credentials
        key: authToken
EOF
logInfo "Username: admin"
logInfo "Password: $ARGOCD_ADMIN_SECRET"
logInfo "Port Forwarding to Argo CD at http://localhost:8443."
logInfo "ctrl-c to abort, happy developing 🧑‍💻🦄"

fg
