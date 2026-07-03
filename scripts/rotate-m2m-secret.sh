#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Logto M2M Secret Rotation Script
# ============================================================
# Purpose: Rotate the client secret of an M2M app
#   1. Add a new secret
#   2. Return the new secret value
#   3. The old secret remains valid until explicitly deleted
#
# Usage:
#   ./scripts/rotate-m2m-secret.sh <m2m-client-id> <logto-mgmt-client-id> <logto-mgmt-client-secret> <logto-endpoint>
# ============================================================

M2M_APP_ID="${1:?Usage: $0 <m2m-app-id> <mgmt-client-id> <mgmt-client-secret> <logto-endpoint>}"
MGMT_CLIENT_ID="${2:?}"
MGMT_CLIENT_SECRET="${3:?}"
LOGTO_ENDPOINT="${4:?}"

echo "==> Getting M2M access token..."
TOKEN_RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/oidc/token" \
  -u "${MGMT_CLIENT_ID}:${MGMT_CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&resource=${LOGTO_ENDPOINT}/api&scope=all")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "ERROR: Failed to get access token. Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "==> Adding new secret to M2M app..."
SECRET_RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "rotated-'"$(date +%s)"'"}')

NEW_SECRET=$(echo "$SECRET_RESPONSE" | jq -r '.value')

if [ -z "$NEW_SECRET" ] || [ "$NEW_SECRET" = "null" ]; then
  echo "ERROR: Failed to create secret. Response: $SECRET_RESPONSE"
  exit 1
fi

echo "==> Secret rotated successfully!"
echo "    New Secret: $NEW_SECRET"
echo ""
echo "Update your .env file with the new secret."
echo "The old secret is still valid. To revoke it:"
echo "  List secrets: GET ${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets"
echo "  Delete old:    DELETE ${LOGTO_ENDPOINT}/api/applications/${M2M_APP_ID}/secrets/<secret-id>"
