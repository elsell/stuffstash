#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

api_image="ghcr.io/example/stuffstash@sha256:$(printf '1%.0s' {1..64})"
web_image="ghcr.io/example/stuffstash-web@sha256:$(printf '2%.0s' {1..64})"
"$repo_root/scripts/build-selfhost-release.sh" \
  --version v0.0.0-test \
  --api-image "$api_image" \
  --web-image "$web_image" \
  --output "$tmp_dir"
(
  cd "$tmp_dir"
  shasum -a 256 -c stuffstash-selfhost.tar.gz.sha256 >/dev/null
)

archive_entries="$(tar -tzf "$tmp_dir/stuffstash-selfhost.tar.gz")"
for required in \
  stuffstash-selfhost/.env.example \
  stuffstash-selfhost/compose.selfhost.yaml \
  stuffstash-selfhost/scripts/selfhost-preflight.sh \
  stuffstash-selfhost/scripts/configure-garage-cors.mjs \
  stuffstash-selfhost/deploy/selfhost/caddy/Caddyfile \
  stuffstash-selfhost/deploy/selfhost/dex/config.yaml \
  stuffstash-selfhost/deploy/selfhost/garage/garage.toml \
  stuffstash-selfhost/deploy/dex/theme/styles.css \
  stuffstash-selfhost/deploy/dex/templates/header.html \
  stuffstash-selfhost/deploy/dex/templates/login.html \
  stuffstash-selfhost/deploy/dex/templates/password.html \
  stuffstash-selfhost/docs/public/brand/stuff-stash-glyph.png \
  stuffstash-selfhost/VERSION; do
  grep -qx "$required" <<<"$archive_entries" || {
    echo "release bundle missing $required" >&2
    exit 1
  }
done

mkdir "$tmp_dir/extracted"
tar -xzf "$tmp_dir/stuffstash-selfhost.tar.gz" -C "$tmp_dir/extracted"
grep -qx "STUFF_STASH_API_IMAGE=$api_image" "$tmp_dir/extracted/stuffstash-selfhost/.env.example"
grep -qx "STUFF_STASH_WEB_IMAGE=$web_image" "$tmp_dir/extracted/stuffstash-selfhost/.env.example"

echo "self-host release bundle tests passed"
