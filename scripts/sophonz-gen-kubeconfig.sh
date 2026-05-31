#!/usr/bin/env bash
# Generate a kubeconfig for a ServiceAccount that has ADMIN rights scoped to the
# `v3` namespace, for use as the `KUBE_CONFIG` GitHub Actions secret.
#
# What it creates in the target cluster:
#   * ServiceAccount        v3/v3-deployer
#   * RoleBinding           v3/v3-deployer-admin  -> ClusterRole `admin` (namespaced)
#   * Secret                v3/v3-deployer-token  (long-lived SA token)
#
# It then prints a self-contained kubeconfig (cluster CA + server + token) whose
# context is named `oci@afextwin`, with the default namespace set to `v3`.
#
# Prerequisites:
#   * kubectl pointed at the target cluster with rights to create the above
#     (run it against your admin context first).
#
# Usage:
#   ./scripts/sophonz-gen-kubeconfig.sh                  # uses current kube-context
#   ./scripts/sophonz-gen-kubeconfig.sh oci@afextwin     # use a specific context
#   ./scripts/sophonz-gen-kubeconfig.sh oci@afextwin > kubeconfig-v3.yaml
#
# Then add the file contents to the repo/org secret KUBE_CONFIG:
#   gh secret set KUBE_CONFIG < kubeconfig-v3.yaml
set -euo pipefail

NAMESPACE="v3"
SA_NAME="v3-deployer"
SECRET_NAME="${SA_NAME}-token"
CONTEXT_NAME="oci@afextwin"
SRC_CONTEXT="${1:-$(kubectl config current-context)}"

log() { echo ">> $*" >&2; }

log "using source context: ${SRC_CONTEXT}"
KUBECTL=(kubectl --context "${SRC_CONTEXT}")

# 1. namespace + ServiceAccount
log "ensuring namespace/${NAMESPACE} and serviceaccount/${SA_NAME}"
"${KUBECTL[@]}" create namespace "${NAMESPACE}" --dry-run=client -o yaml | "${KUBECTL[@]}" apply -f - >/dev/null
"${KUBECTL[@]}" -n "${NAMESPACE}" create serviceaccount "${SA_NAME}" \
  --dry-run=client -o yaml | "${KUBECTL[@]}" apply -f - >/dev/null

# 2. namespace-scoped admin: bind the built-in `admin` ClusterRole via a RoleBinding
log "binding ClusterRole/admin to ${SA_NAME} within namespace ${NAMESPACE}"
cat <<EOF | "${KUBECTL[@]}" apply -f - >/dev/null
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${SA_NAME}-admin
  namespace: ${NAMESPACE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
  - kind: ServiceAccount
    name: ${SA_NAME}
    namespace: ${NAMESPACE}
EOF

# 3. long-lived token Secret bound to the SA (token does not expire, unlike
#    `kubectl create token` which is capped). Suited for a stored CI secret.
log "creating long-lived token secret ${SECRET_NAME}"
cat <<EOF | "${KUBECTL[@]}" apply -f - >/dev/null
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
  namespace: ${NAMESPACE}
  annotations:
    kubernetes.io/service-account.name: ${SA_NAME}
type: kubernetes.io/service-account-token
EOF

# wait for the token controller to populate the secret
log "waiting for token to be populated"
for _ in $(seq 1 30); do
  TOKEN="$("${KUBECTL[@]}" -n "${NAMESPACE}" get secret "${SECRET_NAME}" -o jsonpath='{.data.token}' 2>/dev/null || true)"
  [[ -n "${TOKEN}" ]] && break
  sleep 1
done
if [[ -z "${TOKEN:-}" ]]; then
  echo "ERROR: token was not populated in ${NAMESPACE}/${SECRET_NAME}" >&2
  exit 1
fi
TOKEN="$(echo "${TOKEN}" | base64 -d)"

# 4. pull cluster server + CA from the source context
CLUSTER_NAME="$("${KUBECTL[@]}" config view -o jsonpath="{.contexts[?(@.name=='${SRC_CONTEXT}')].context.cluster}")"
SERVER="$("${KUBECTL[@]}" config view -o jsonpath="{.clusters[?(@.name=='${CLUSTER_NAME}')].cluster.server}")"
CA_DATA="$("${KUBECTL[@]}" config view --raw -o jsonpath="{.clusters[?(@.name=='${CLUSTER_NAME}')].cluster.certificate-authority-data}")"

if [[ -z "${CA_DATA}" ]]; then
  # certificate-authority is a file path rather than inline data; inline it.
  CA_FILE="$("${KUBECTL[@]}" config view --raw -o jsonpath="{.clusters[?(@.name=='${CLUSTER_NAME}')].cluster.certificate-authority}")"
  if [[ -n "${CA_FILE}" && -f "${CA_FILE}" ]]; then
    CA_DATA="$(base64 < "${CA_FILE}" | tr -d '\n')"
  fi
fi

log "server=${SERVER}"

# 5. emit the kubeconfig (context renamed to ${CONTEXT_NAME})
cat <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: afextwin
    cluster:
      server: ${SERVER}
      certificate-authority-data: ${CA_DATA}
contexts:
  - name: ${CONTEXT_NAME}
    context:
      cluster: afextwin
      namespace: ${NAMESPACE}
      user: ${SA_NAME}
current-context: ${CONTEXT_NAME}
users:
  - name: ${SA_NAME}
    user:
      token: ${TOKEN}
EOF

log "done. context '${CONTEXT_NAME}' -> namespace '${NAMESPACE}' as ${SA_NAME} (admin)."
log "store the stdout above as the KUBE_CONFIG secret."
