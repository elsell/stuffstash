#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

env_file="$tmp_dir/.env"
dex_config="$tmp_dir/dex/config.yaml"

"$repo_root/scripts/configure-selfhost.sh" \
  --host 192.168.2.52 \
  --env-file "$env_file" \
  --dex-config "$dex_config" >"$tmp_dir/output"

grep -qx 'STUFF_STASH_BIND_ADDRESS=0.0.0.0' "$env_file"
grep -qx 'STUFF_STASH_SELFHOST_HOSTNAME=192.168.2.52' "$env_file"
grep -qx 'STUFF_STASH_WEB_ORIGIN=https://192.168.2.52:8081' "$env_file"
grep -qx 'STUFF_STASH_API_ORIGIN=https://192.168.2.52:8080' "$env_file"
grep -qx 'STUFF_STASH_OIDC_ISSUER=https://192.168.2.52:5556/dex' "$env_file"
grep -qx "DEX_CONFIG_PATH=$dex_config" "$env_file"
grep -qx 'issuer: https://192.168.2.52:5556/dex' "$dex_config"
grep -qx '    - https://192.168.2.52:8081' "$dex_config"
grep -q 'Open https://192.168.2.52:8081' "$tmp_dir/output"

mode="$(stat -c '%a' "$dex_config" 2>/dev/null || stat -f '%Lp' "$dex_config")"
[ "$mode" = "600" ] || {
  echo "configured Dex file mode was $mode, want 600" >&2
  exit 1
}

if "$repo_root/scripts/configure-selfhost.sh" \
  --host 192.168.2.999 \
  --env-file "$tmp_dir/invalid.env" \
  --dex-config "$tmp_dir/invalid-dex.yaml" >/dev/null 2>&1; then
  echo "self-host setup accepted an invalid IPv4 address" >&2
  exit 1
fi

if "$repo_root/scripts/configure-selfhost.sh" \
  --host 192.168.2.52 \
  --env-file "$env_file" \
  --dex-config "$dex_config" >/dev/null 2>&1; then
  echo "self-host setup overwrote an existing configuration" >&2
  exit 1
fi

mkdir "$tmp_dir/bin"
cat > "$tmp_dir/bin/ip" <<'EOF'
#!/usr/bin/env bash
if [ "$1 $2 $3" = "-4 route get" ]; then
  echo "1.1.1.1 dev tun0 src 100.64.0.2"
else
  echo "2: eth0 inet 192.168.2.52/24 brd 192.168.2.255 scope global eth0"
  echo "3: tun0 inet 100.64.0.2/32 scope global tun0"
fi
EOF
chmod +x "$tmp_dir/bin/ip"
PATH="$tmp_dir/bin:$PATH" "$repo_root/scripts/configure-selfhost.sh" \
  --env-file "$tmp_dir/detected.env" \
  --dex-config "$tmp_dir/detected-dex.yaml" >"$tmp_dir/detected-output"
grep -qx 'STUFF_STASH_SELFHOST_HOSTNAME=192.168.2.52' "$tmp_dir/detected.env"

echo "self-host configuration tests passed"
