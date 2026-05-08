#!/usr/bin/env bash
# Apache License 2.0
# Copyright 2026 external-dns-technitium-webhook Contributors
#
# End-to-end smoke test: verify that external-dns + this webhook +
# the Bugs5382/helm-technitium-chart can publish a DNS record into a
# Technitium zone. Designed to run in GitHub Actions on a kind cluster
# but should also work against any cluster the caller's KUBECONFIG
# points at.
#
# Required env (with sane defaults):
#   TECHNITIUM_CHART_PATH  Path to the local technitium chart directory
#   ZONE                   Zone to create (default: example.test)
#   RECORD                 FQDN we expect external-dns to publish
#   RECORD_TARGET          Expected A-record value
#   ADMIN_USER / ADMIN_PASS  Technitium admin credentials

set -euo pipefail

TECHNITIUM_CHART_PATH="${TECHNITIUM_CHART_PATH:-./helm-technitium-chart/technitium}"
TECHNITIUM_NAMESPACE="${TECHNITIUM_NAMESPACE:-technitium}"
EXTERNAL_DNS_NAMESPACE="${EXTERNAL_DNS_NAMESPACE:-external-dns}"
ZONE="${ZONE:-example.test}"
RECORD="${RECORD:-e2e.example.test}"
RECORD_TARGET="${RECORD_TARGET:-10.0.0.42}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-admin}"
WEBHOOK_IMAGE="${WEBHOOK_IMAGE:-external-dns-technitium-webhook:e2e}"
EXTERNAL_DNS_CHART_VERSION="${EXTERNAL_DNS_CHART_VERSION:-1.19.0}"

HERE="$(cd "$(dirname "$0")" && pwd)"

PF_PID=""
cleanup_pf() { [[ -n "$PF_PID" ]] && kill "$PF_PID" 2>/dev/null || true; PF_PID=""; }
trap cleanup_pf EXIT

log()  { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
fail() { printf '\n\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

require() { command -v "$1" >/dev/null || fail "missing required tool: $1"; }
require kubectl; require helm; require curl; require jq

# Port-forward 5380 → localhost:5380 in the background, wait until reachable.
port_forward_technitium() {
  cleanup_pf
  kubectl -n "$TECHNITIUM_NAMESPACE" port-forward svc/technitium 5380:5380 \
    >/tmp/pf.log 2>&1 &
  PF_PID=$!
  for _ in $(seq 1 30); do
    if curl -sSf "http://127.0.0.1:5380/" >/dev/null 2>&1 \
       || curl -sS  "http://127.0.0.1:5380/api/user/login" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  fail "Technitium port-forward never became reachable. Logs:\n$(cat /tmp/pf.log)"
}

# Authenticate against Technitium and echo the session token.
login() {
  local resp token
  resp=$(curl -sS --get \
    --data-urlencode "user=$ADMIN_USER" \
    --data-urlencode "pass=$ADMIN_PASS" \
    "http://127.0.0.1:5380/api/user/login")
  token=$(echo "$resp" | jq -r '.token // empty')
  [[ -n "$token" ]] || fail "login failed; response: $resp"
  echo "$token"
}

create_zone() {
  local token="$1"
  log "Creating primary zone $ZONE"
  local resp
  resp=$(curl -sS --get \
    --data-urlencode "token=$token" \
    --data-urlencode "zone=$ZONE" \
    --data-urlencode "type=Primary" \
    "http://127.0.0.1:5380/api/zones/create")
  local status; status=$(echo "$resp" | jq -r '.status // empty')
  if [[ "$status" != "ok" ]]; then
    # Already exists is fine for re-runs.
    if echo "$resp" | grep -qi "already exists"; then
      log "Zone $ZONE already exists, continuing"
      return 0
    fi
    fail "zone create failed: $resp"
  fi
}

# Poll the zone for the expected record. external-dns syncs every 30s in
# our values, so 3 minutes of slack is plenty.
wait_for_record() {
  local token="$1"
  log "Waiting for $RECORD → $RECORD_TARGET in zone $ZONE"
  local resp
  for i in $(seq 1 36); do
    resp=$(curl -sS --get \
      --data-urlencode "token=$token" \
      --data-urlencode "domain=$ZONE" \
      --data-urlencode "zone=$ZONE" \
      --data-urlencode "listZone=true" \
      "http://127.0.0.1:5380/api/zones/records/get")
    if echo "$resp" \
        | jq -e --arg n "$RECORD" --arg ip "$RECORD_TARGET" '
            (.response.records // [])
            | map(select(
                (.name | ascii_downcase) == ($n | ascii_downcase)
                and .type == "A"
                and (.rData.ipAddress == $ip or .rData.value == $ip)
              ))
            | length > 0
          ' >/dev/null; then
      log "Record found on attempt $i:"
      echo "$resp" | jq --arg n "$RECORD" '
        (.response.records // [])
        | map(select((.name | ascii_downcase) == ($n | ascii_downcase)))
      '
      return 0
    fi
    printf '.'
    sleep 5
  done
  echo
  echo "Last zone listing:"
  echo "$resp" | jq . || echo "$resp"
  fail "record $RECORD did not appear in zone $ZONE within timeout"
}

dump_diagnostics() {
  echo "::group::Pods"
  kubectl get pods -A -o wide || true
  echo "::endgroup::"
  echo "::group::external-dns describe"
  kubectl -n "$EXTERNAL_DNS_NAMESPACE" describe pod -l app.kubernetes.io/name=external-dns || true
  echo "::endgroup::"
  echo "::group::external-dns container logs"
  kubectl -n "$EXTERNAL_DNS_NAMESPACE" logs -l app.kubernetes.io/name=external-dns -c external-dns --tail=300 || true
  echo "::endgroup::"
  echo "::group::webhook sidecar logs"
  kubectl -n "$EXTERNAL_DNS_NAMESPACE" logs -l app.kubernetes.io/name=external-dns -c webhook --tail=300 || true
  echo "::endgroup::"
  echo "::group::Technitium logs"
  kubectl -n "$TECHNITIUM_NAMESPACE" logs deploy/technitium --tail=300 || true
  echo "::endgroup::"
}
trap 'rc=$?; if [[ $rc -ne 0 ]]; then dump_diagnostics; fi; cleanup_pf; exit $rc' EXIT

###############################################################################
# 1. Install Technitium
###############################################################################
log "Installing Technitium chart from $TECHNITIUM_CHART_PATH"
kubectl create namespace "$TECHNITIUM_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install technitium "$TECHNITIUM_CHART_PATH" \
  --namespace "$TECHNITIUM_NAMESPACE" \
  -f "$HERE/values-technitium.yaml" \
  --wait --timeout 5m

log "Waiting for Technitium pod readiness"
kubectl -n "$TECHNITIUM_NAMESPACE" rollout status deploy/technitium --timeout=300s

###############################################################################
# 2. Bootstrap the test zone via the Technitium API
###############################################################################
port_forward_technitium
TOKEN="$(login)"
create_zone "$TOKEN"
cleanup_pf

###############################################################################
# 3. Install external-dns + this webhook
###############################################################################
log "Installing external-dns with webhook sidecar"
kubectl create namespace "$EXTERNAL_DNS_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$EXTERNAL_DNS_NAMESPACE" create secret generic technitium-credentials \
  --from-literal=username="$ADMIN_USER" \
  --from-literal=password="$ADMIN_PASS" \
  --dry-run=client -o yaml | kubectl apply -f -

helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/ >/dev/null
helm repo update external-dns >/dev/null
helm upgrade --install external-dns external-dns/external-dns \
  --version "$EXTERNAL_DNS_CHART_VERSION" \
  --namespace "$EXTERNAL_DNS_NAMESPACE" \
  -f "$HERE/values-external-dns.yaml" \
  --wait --timeout 5m

###############################################################################
# 4. Apply the workload that should produce a DNS record
###############################################################################
log "Applying workload"
kubectl apply -f "$HERE/workload.yaml"

###############################################################################
# 5. Verify the record landed in Technitium
###############################################################################
port_forward_technitium
TOKEN="$(login)"
wait_for_record "$TOKEN"
cleanup_pf

log "✅ e2e PASSED — $RECORD published to zone $ZONE"
