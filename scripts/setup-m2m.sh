#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Logto M2M Bootstrap Script
# ============================================================
# Purpose:
#   1. Create an M2M app in Logto via Management API (using Logto CLI token)
#   2. Assign a custom role with only POST /api/users permission
#   3. Output credentials to be saved in .env
#
# Prerequisites:
#   - Logto running (npm start in logto container)
#   - Logto CLI available: npx @logto/cli
#   - Logto admin credentials (from first-run setup)
#
# Usage:
#   ./scripts/setup-m2m.sh <admin-username> <admin-password> <logto-endpoint>
# ============================================================

ADMIN_USERNAME="${1:?Usage: $0 <admin-username> <admin-password> <logto-endpoint>}"
ADMIN_PASSWORD="${2:?}"
LOGTO_ENDPOINT="${3:?}"

echo "==> Logging into Logto CLI..."
TOKEN=$(npx @logto/cli@latest token add -e "$LOGTO_ENDPOINT" -u "$ADMIN_USERNAME" -p "$ADMIN_PASSWORD" 2>/dev/null)

echo "==> Creating M2M app 'finsplitter-m2m'..."
RESPONSE=$(curl -s -X POST "${LOGTO_ENDPOINT}/api/applications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "finsplitter-m2m",
    "type": "MachineToMachine",
    "description": "Finsplitter backend M2M app for Management API access"
  }')

APP_ID=$(echo "$RESPONSE" | jq -r '.id')
APP_SECRET=$(echo "$RESPONSE" | jq -r '.secret')

if [ -z "$APP_ID" ] || [ "$APP_ID" = "null" ]; then
  echo "ERROR: Failed to create M2M app. Response: $RESPONSE"
  exit 1
fi

echo "==> M2M app created!"
echo "    Client ID:  $APP_ID"
echo "    Client Secret: $APP_SECRET"
echo ""
echo "Add these to your .env file:"
echo "  LOGTO_MGMT_CLIENT_ID=$APP_ID"
echo "  LOGTO_MGMT_CLIENT_SECRET=$APP_SECRET"
echo ""
echo "NOTE: The custom role (only POST /api/users) must be assigned manually in Logto Console:"
echo "  1. Go to Console > Machine-to-machine > finsplitter-m2m"
echo "  2. Go to 'Roles' tab"
echo "  3. Assign the custom 'finsplitter-m2m-users-only' role (create it with scope: create:users)"
echo ""
echo "Then run: ./scripts/rotate-m2m-secret.sh $APP_ID"
