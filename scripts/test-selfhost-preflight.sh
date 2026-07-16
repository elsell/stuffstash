#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

cp "$repo_root/.env.example" "$tmp_dir/.env"
printf '%s\n' "PREFLIGHT_INJECTION=\$(touch $tmp_dir/injected)" >> "$tmp_dir/.env"

SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >/dev/null
[ ! -e "$tmp_dir/injected" ] || {
  echo "preflight executed shell content from .env" >&2
  exit 1
}

if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted unchanged example secrets" >&2
  exit 1
fi

sed -i.bak \
  -e 's/change-me-postgres/test-postgres/' \
  -e 's/change-me-spicedb-postgres/test-spicedb-postgres/' \
  -e 's/change-me-spicedb-preshared-key/test-spicedb-key/' \
  -e 's/change-me-garage-access-key/test-garage-access/' \
  -e 's/change-me-garage-secret-key-change-me-garage-secret-key/test-garage-secret/' \
  -e 's|AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=|MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=|' \
  "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >"$tmp_dir/dex-output" 2>&1; then
  echo "strict preflight accepted the bundled Dex identities" >&2
  exit 1
fi
grep -q 'private Dex config' "$tmp_dir/dex-output"

cp "$repo_root/deploy/selfhost/dex/config.yaml" "$tmp_dir/dex-private.yaml"
sed -i.bak \
  -e 's/owner@example\.com/alice@example.test/g' \
  -e 's/viewer@example\.com/bob@example.test/g' \
  -e 's/stuff-stash-local-secret/test-private-client-secret/g' \
  "$tmp_dir/dex-private.yaml"
chmod 600 "$tmp_dir/dex-private.yaml"
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-private.yaml|" "$tmp_dir/.env"
SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null

cp "$tmp_dir/dex-private.yaml" "$tmp_dir/dex-comment-bypass.yaml"
sed -i.bak 's|^issuer: .*|issuer: https://evil.example/dex|' "$tmp_dir/dex-comment-bypass.yaml"
printf '%s\n' '# issuer: https://stuffstash.localhost:5556/dex' >> "$tmp_dir/dex-comment-bypass.yaml"
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-comment-bypass.yaml|" "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted Dex values found only in comments" >&2
  exit 1
fi
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-private.yaml|" "$tmp_dir/.env"

awk '
  /- id: stuff-stash-web-local/ { in_web=1 }
  in_web && /- https:\/\/stuffstash\.localhost:8081\/callback/ {
    sub(/https:\/\/stuffstash\.localhost:8081\/callback/, "https://evil.example/callback")
    in_web=0
  }
  { print }
' "$tmp_dir/dex-private.yaml" > "$tmp_dir/dex-wrong-web-redirect.yaml"
chmod 600 "$tmp_dir/dex-wrong-web-redirect.yaml"
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-wrong-web-redirect.yaml|" "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted a redirect on the wrong Dex client" >&2
  exit 1
fi
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-private.yaml|" "$tmp_dir/.env"

sed -i.bak 's/stuffstash\.localhost/192.168.2.52/g' "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >"$tmp_dir/output" 2>&1; then
  echo "preflight accepted an IP-literal OIDC hostname" >&2
  exit 1
fi
grep -q 'DNS hostname' "$tmp_dir/output"

cp "$repo_root/.env.example" "$tmp_dir/.env"
sed -i.bak 's/^STUFF_STASH_BIND_ADDRESS=.*/STUFF_STASH_BIND_ADDRESS=0.0.0.0/' "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "trial preflight allowed example credentials on a non-loopback bind" >&2
  exit 1
fi

sed -i.bak \
  -e 's/^STUFF_STASH_BIND_ADDRESS=.*/STUFF_STASH_BIND_ADDRESS=127.0.0.1/' \
  -e 's|^STUFF_STASH_OIDC_ISSUER=.*|STUFF_STASH_OIDC_ISSUER=https://stuffstash.localhost:5556@evil.example/dex|' \
  "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "preflight accepted a URL userinfo hostname bypass" >&2
  exit 1
fi

echo "self-host preflight tests passed"
