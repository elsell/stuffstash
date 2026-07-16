#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

"$repo_root/scripts/build-selfhost-release.sh" --version v0.0.0-test --output "$tmp_dir"
(
  cd "$tmp_dir"
  shasum -a 256 -c stuffstash-selfhost.tar.gz.sha256 >/dev/null
)

archive_entries="$(tar -tzf "$tmp_dir/stuffstash-selfhost.tar.gz")"
for required in \
  stuffstash-selfhost/.env.example \
  stuffstash-selfhost/compose.selfhost.yaml \
  stuffstash-selfhost/scripts/selfhost-preflight.sh \
  stuffstash-selfhost/deploy/selfhost/caddy/Caddyfile \
  stuffstash-selfhost/deploy/selfhost/dex/config.yaml \
  stuffstash-selfhost/VERSION; do
  grep -qx "$required" <<<"$archive_entries" || {
    echo "release bundle missing $required" >&2
    exit 1
  }
done

echo "self-host release bundle tests passed"
