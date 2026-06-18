#!/usr/bin/env bash
# Refresh the standing demo's LIVE evidence.
#
# Applies the showcase against real Azure DevOps, then GETs each created resource
# from the live REST API and writes the responses to ./evidence/. Leave the demo
# standing for review; destroy only on your go-ahead (`terraform destroy`).
#
# Prereqs: secrets.env populated (AZDO_ORG_SERVICE_URL, AZDO_PERSONAL_ACCESS_TOKEN),
# plus a terraform.tfvars (copy terraform.tfvars.example). Run from this directory.
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
cd "$REPO"

# 1. Credentials.
[ -f secrets.env ] || { echo "FATAL: secrets.env not found at repo root" >&2; exit 1; }
set -a; . ./secrets.env; set +a
: "${AZDO_ORG_SERVICE_URL:?set in secrets.env}"
: "${AZDO_PERSONAL_ACCESS_TOKEN:?set in secrets.env}"

# 2. Build the provider + point Terraform at it via dev_overrides.
echo "==> building provider"
go build -mod=vendor -o /tmp/tf-betterado-devbin/terraform-provider-betterado .
TFRC="$(mktemp)"
cat > "$TFRC" <<EOF
provider_installation {
  dev_overrides { "parsoFish/betterado" = "/tmp/tf-betterado-devbin" }
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE="$TFRC"

cd "$HERE"
mkdir -p evidence

# 3. Apply, then prove idempotency (a clean re-plan = round-trip is honest).
echo "==> terraform apply"
terraform apply -auto-approve
echo "==> idempotency re-plan (must be: No changes)"
terraform plan -detailed-exitcode || { [ $? -eq 2 ] && { echo "FAIL: perpetual diff — a flatten path does not round-trip" >&2; exit 1; }; }

# 4. Capture live REST evidence (the round-trip proof, not a test-name table).
ORG="${AZDO_ORG_SERVICE_URL%/}"
VSRM="${ORG/https:\/\/dev.azure.com/https://vsrm.dev.azure.com}"
PROJECT="$(terraform output -raw release_definition_url | sed -E 's#.*/([^/]+)/_release.*#\1#')"
RDID="$(terraform output -raw release_definition_id)"
TGID="$(terraform output -raw task_group_id)"
AUTH=(-u ":${AZDO_PERSONAL_ACCESS_TOKEN}")

echo "==> GET release definition $RDID (vsrm host)"
curl -sf "${AUTH[@]}" "${VSRM}/${PROJECT}/_apis/release/definitions/${RDID}?api-version=7.1" \
  -o evidence/release-definition.json
echo "==> GET release folders"
curl -sf "${AUTH[@]}" "${VSRM}/${PROJECT}/_apis/release/folders?api-version=7.1" \
  -o evidence/release-folders.json
echo "==> GET task group $TGID (core host)"
curl -sf "${AUTH[@]}" "${ORG}/${PROJECT}/_apis/distributedtask/taskgroups/${TGID}?api-version=7.1" \
  -o evidence/task-group.json

echo
echo "Live evidence written to demo/standing/evidence/. The resources are STANDING"
echo "in ADO — review them in the portal, then 'terraform destroy' when done."
