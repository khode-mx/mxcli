#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# auth-discovery-spike.sh — Map the Mendix Marketplace API surface.
#
# Round-3 scope. We have already confirmed (round 2) that PAT via
# "Authorization: MxToken <pat>" works against marketplace-api.mendix.com.
# This round discovers endpoint paths and response shapes so we can design
# the marketplace package.
#
# Reads $MENDIX_PAT from the environment; never logs it.
# Output: /tmp/auth-spike-report.md (redacted — safe to share).
#
# Usage:
#   export MENDIX_PAT='<your-pat>'     # create at https://user-settings.mendix.com/
#   ./scripts/auth-discovery-spike.sh
#   cat /tmp/auth-spike-report.md

set -u

REPORT=/tmp/auth-spike-report.md
MARKETPLACE="https://marketplace-api.mendix.com"
CATALOG="https://catalog.mendix.com"

if [[ -z "${MENDIX_PAT:-}" ]]; then
  echo "error: MENDIX_PAT is not set" >&2
  echo "create a PAT at https://user-settings.mendix.com/ and export it:" >&2
  echo "  export MENDIX_PAT='...'" >&2
  exit 2
fi

# Redact the PAT from any output we write to the report.
redact() {
  sed "s|${MENDIX_PAT}|MENDIX_PAT_REDACTED|g"
}

# probe <label> <method> <url> [curl args...]
probe() {
  local label="$1"
  local method="$2"
  local url="$3"
  shift 3

  local tmp_body tmp_headers code
  tmp_body=$(mktemp)
  tmp_headers=$(mktemp)
  code=$(curl -sS -o "$tmp_body" -D "$tmp_headers" -w "%{http_code}" \
    -X "$method" \
    -H "Authorization: MxToken $MENDIX_PAT" \
    -H "Accept: application/json" \
    "$@" \
    "$url" 2>&1) || code="CURL_ERR"

  echo "### $label"
  echo
  echo '```'
  echo "$method $url"
  echo "HTTP $code"
  echo "--- response headers ---"
  head -20 "$tmp_headers" | redact
  echo "--- response body (first 60 lines) ---"
  head -60 "$tmp_body" | redact
  echo '```'
  echo

  rm -f "$tmp_body" "$tmp_headers"
}

{
  echo "# Marketplace API Spike Report"
  echo
  echo "Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo "Marketplace base: \`$MARKETPLACE\`"
  echo "Catalog base:     \`$CATALOG\`"
  echo "Auth scheme:      \`Authorization: MxToken <pat>\`"
  echo
  echo "PAT is redacted as \`MENDIX_PAT_REDACTED\`. Safe to share."
  echo

  echo "## Marketplace API — endpoint discovery"
  echo
  echo "Known working: \`GET /v1/content\`. Map out the rest of the surface."
  echo

  probe "GET /v1/content (known working)"     GET "$MARKETPLACE/v1/content?limit=3"
  probe "GET /v1/content/2888 (component detail — DB Connector)" \
                                              GET "$MARKETPLACE/v1/content/2888"
  probe "GET /v1/content?search=database"     GET "$MARKETPLACE/v1/content?search=database&limit=3"
  probe "GET /v1/content/2888/versions"       GET "$MARKETPLACE/v1/content/2888/versions"
  probe "GET /v1/content/2888/releases"       GET "$MARKETPLACE/v1/content/2888/releases"
  probe "GET /v1 (index — hope for hypermedia)" \
                                              GET "$MARKETPLACE/v1"
  probe "GET / (root — hope for hypermedia)"  GET "$MARKETPLACE/"

  echo "## Catalog API — sanity check"
  echo
  echo "Same PAT against catalog.mendix.com should 200 (or at least not 401)."
  echo

  probe "GET catalog /rest/catalog/v3/entities" \
                                              GET "$CATALOG/rest/catalog/v3/entities?limit=1"
  probe "GET catalog /v1/register (doc example)" \
                                              GET "$CATALOG/v1/register"
  probe "GET catalog /"                       GET "$CATALOG/"

  echo "## Invalid PAT — confirms ErrUnauthenticated shape"
  echo
  probe "GET /v1/content with bad PAT" \
    GET "$MARKETPLACE/v1/content?limit=1" \
    -H "Authorization: MxToken obviously-not-a-real-token"
  # Note: second -H "Authorization:" overrides the first; that's what we want
  # to test, but curl's -H handling sends both. Use --header to replace? Actually
  # curl uses the last occurrence, which is what we want here.

  echo "## Summary — fill in after running"
  echo
  echo "- [ ] \`/v1/content\` (list) returns 200 and list-shaped JSON"
  echo "- [ ] \`/v1/content/{id}\` (detail) returns 200 and describes one module"
  echo "- [ ] \`?search=…\` parameter is accepted"
  echo "- [ ] Versions are available: which path? inline in detail?"
  echo "- [ ] Download URL for .mpk: which path?"
  echo "- [ ] Bad PAT returns: 401 / 403 / other?"
  echo
  echo "**Next step**: feed findings into \`docs/11-proposals/PROPOSAL_marketplace_modules.md\`"
  echo "to correct the endpoint list and base URL."
} > "$REPORT"

echo "report written to $REPORT"
echo "  (PAT redacted — safe to paste findings into proposals)"
