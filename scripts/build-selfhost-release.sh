#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version=""
output="$repo_root/dist"

fail() {
  echo "self-host bundle failed: $*" >&2
  exit 1
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version) shift; version="${1:-}" ;;
    --output) shift; output="${1:-}" ;;
    *) fail "unknown option: $1" ;;
  esac
  shift
done
[ -n "$version" ] || fail "--version is required"
[ -n "$output" ] || fail "--output is required"

stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
bundle="$stage/stuffstash-selfhost"
mkdir -p \
  "$bundle/deploy/selfhost/caddy" \
  "$bundle/deploy/selfhost/dex" \
  "$bundle/deploy/selfhost/garage" \
  "$bundle/deploy/dex/theme" \
  "$bundle/deploy/dex/templates" \
  "$bundle/docs/public/brand" \
  "$bundle/scripts" \
  "$output"

cp "$repo_root/.env.example" "$bundle/.env.example"
cp "$repo_root/compose.selfhost.yaml" "$bundle/compose.selfhost.yaml"
cp "$repo_root/deploy/selfhost/caddy/Caddyfile" "$bundle/deploy/selfhost/caddy/Caddyfile"
cp "$repo_root/deploy/selfhost/dex/config.yaml" "$bundle/deploy/selfhost/dex/config.yaml"
cp "$repo_root/deploy/selfhost/garage/garage.toml" "$bundle/deploy/selfhost/garage/garage.toml"
cp "$repo_root/deploy/dex/theme/styles.css" "$bundle/deploy/dex/theme/styles.css"
cp "$repo_root/deploy/dex/templates/header.html" "$bundle/deploy/dex/templates/header.html"
cp "$repo_root/deploy/dex/templates/login.html" "$bundle/deploy/dex/templates/login.html"
cp "$repo_root/deploy/dex/templates/password.html" "$bundle/deploy/dex/templates/password.html"
cp "$repo_root/docs/public/brand/stuff-stash-glyph.png" "$bundle/docs/public/brand/stuff-stash-glyph.png"
cp "$repo_root/scripts/configure-garage-cors.mjs" "$bundle/scripts/configure-garage-cors.mjs"
cp "$repo_root/scripts/selfhost-preflight.sh" "$bundle/scripts/selfhost-preflight.sh"
printf '%s\n' "$version" > "$bundle/VERSION"

archive="$output/stuffstash-selfhost.tar.gz"
tar -C "$stage" -czf "$archive" stuffstash-selfhost
(
  cd "$output"
  shasum -a 256 stuffstash-selfhost.tar.gz > stuffstash-selfhost.tar.gz.sha256
)

echo "Built $archive"
