#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

cp "$repo_root/.env.example" "$tmp_dir/.env"
printf '%s\n' "PREFLIGHT_INJECTION=\$(touch $tmp_dir/injected)" >> "$tmp_dir/.env"

SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null
[ ! -e "$tmp_dir/injected" ] || {
  echo "preflight executed shell content from .env" >&2
  exit 1
}

if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
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
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >"$tmp_dir/dex-output" 2>&1; then
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
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted the bundled Dex password hash" >&2
  exit 1
fi

sed -i.bak \
  -e 's|[$]2a[$]10[$]2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W|$2a$10$e6Kyp/nZqCGBTYEYLBeEy.47O8wZ0ncGjdJSf18fb0KkwzCZXoyGO|g' \
  -e 's/11111111-1111-1111-1111-111111111111/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/g' \
  -e 's/22222222-2222-2222-2222-222222222222/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb/g' \
  "$tmp_dir/dex-private.yaml"
SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null

cp "$tmp_dir/dex-private.yaml" "$tmp_dir/dex-comment-bypass.yaml"
sed -i.bak 's|^issuer: .*|issuer: https://evil.example/dex|' "$tmp_dir/dex-comment-bypass.yaml"
printf '%s\n' '# issuer: https://stuffstash.localhost:5556/dex' >> "$tmp_dir/dex-comment-bypass.yaml"
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-comment-bypass.yaml|" "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
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
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted a redirect on the wrong Dex client" >&2
  exit 1
fi
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-private.yaml|" "$tmp_dir/.env"

sed -i.bak \
  -e 's/stuffstash\.localhost/192.168.2.52/g' \
  -e 's/^STUFF_STASH_BIND_ADDRESS=.*/STUFF_STASH_BIND_ADDRESS=0.0.0.0/' \
  "$tmp_dir/.env"
sed -i.bak 's/stuffstash\.localhost/192.168.2.52/g' "$tmp_dir/dex-private.yaml"
SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null

for malformed_hostname in \
  'good.example@evil.example' \
  'good.example:444' \
  'good..example' \
  '-good.example' \
  'good-.example' \
  'good example' \
  '192.168.2.999' \
  '192.168.2' \
  '192.168.02.52' \
  '0.0.0.0'; do
  cp "$repo_root/.env.example" "$tmp_dir/malformed-host.env"
  sed -i.bak "s|stuffstash\.localhost|$malformed_hostname|g" "$tmp_dir/malformed-host.env"
  if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
    "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/malformed-host.env" >/dev/null 2>&1; then
    echo "preflight accepted malformed hostname: $malformed_hostname" >&2
    exit 1
  fi
done

cp "$repo_root/.env.example" "$tmp_dir/.env"
sed -i.bak \
  -e 's/stuffstash\.localhost/192.168.2.52/g' \
  -e 's/^STUFF_STASH_BIND_ADDRESS=.*/STUFF_STASH_BIND_ADDRESS=0.0.0.0/' \
  "$tmp_dir/.env"
cp "$repo_root/deploy/selfhost/dex/config.yaml" "$tmp_dir/dex-example-ip.yaml"
sed -i.bak 's/stuffstash\.localhost/192.168.2.52/g' "$tmp_dir/dex-example-ip.yaml"
chmod 600 "$tmp_dir/dex-example-ip.yaml"
sed -i.bak "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-example-ip.yaml|" "$tmp_dir/.env"
SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >"$tmp_dir/output" 2>&1
grep -q 'Anyone who can reach this server can use the example credentials' "$tmp_dir/output"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --strict --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight allowed example credentials" >&2
  exit 1
fi

sed -i.bak \
  -e 's/^STUFF_STASH_BIND_ADDRESS=.*/STUFF_STASH_BIND_ADDRESS=127.0.0.1/' \
  -e 's|^STUFF_STASH_OIDC_ISSUER=.*|STUFF_STASH_OIDC_ISSUER=https://stuffstash.localhost:5556@evil.example/dex|' \
  "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "preflight accepted a URL userinfo hostname bypass" >&2
  exit 1
fi

echo "self-host preflight tests passed"
