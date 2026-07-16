#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
host=""
env_file="$repo_root/.env"
dex_config="$repo_root/.stuffstash/selfhost/dex/config.yaml"
dex_config_value="./.stuffstash/selfhost/dex/config.yaml"

usage() {
  echo "Usage: scripts/configure-selfhost.sh [--host IPV4] [--env-file PATH] [--dex-config PATH]"
}

fail() {
  echo "Setup failed: $*" >&2
  exit 1
}

valid_ipv4() {
  local address="$1" octet
  local octets=()
  [[ "$address" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] || return 1
  [ "$address" != "0.0.0.0" ] && [ "$address" != "127.0.0.1" ] || return 1
  IFS='.' read -r -a octets <<< "$address"
  for octet in "${octets[@]}"; do
    [[ "$octet" == "0" || "$octet" =~ ^[1-9][0-9]*$ ]] || return 1
    [ "$octet" -le 255 ] || return 1
  done
}

private_ipv4() {
  local address="$1" second
  case "$address" in
    10.*|192.168.*) return 0 ;;
    172.*)
      second="${address#172.}"
      second="${second%%.*}"
      [ "$second" -ge 16 ] && [ "$second" -le 31 ]
      ;;
    *) return 1 ;;
  esac
}

detect_host() {
  local candidate="" interface="" candidates=""
  if command -v ip >/dev/null 2>&1; then
    candidate="$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{ for (i=1; i<=NF; i++) if ($i == "src") { print $(i+1); exit } }')"
    if valid_ipv4 "$candidate" && private_ipv4 "$candidate"; then
      printf '%s\n' "$candidate"
      return
    fi
    candidates="$(ip -o -4 addr show scope global 2>/dev/null | awk '$2 !~ /^(lo|tun|tap|wg|tailscale)/ { split($4, address, "/"); print address[1] }')"
  fi
  if command -v route >/dev/null 2>&1 && command -v ipconfig >/dev/null 2>&1; then
    interface="$(route -n get default 2>/dev/null | awk '/interface:/ { print $2; exit }')"
    [ -z "$interface" ] || candidate="$(ipconfig getifaddr "$interface" 2>/dev/null || true)"
    if valid_ipv4 "$candidate" && private_ipv4 "$candidate"; then
      printf '%s\n' "$candidate"
      return
    fi
  fi
  if command -v ifconfig >/dev/null 2>&1; then
    candidates="$candidates
$(ifconfig 2>/dev/null | awk '
  /^[^[:space:]]/ { interface=$1; sub(/:$/, "", interface) }
  /^[[:space:]]*inet / && interface !~ /^(lo|utun|tun|tap|gif|stf|bridge|awdl|llw)/ { print $2 }
')"
  fi
  if command -v hostname >/dev/null 2>&1; then
    candidates="$candidates
$(hostname -I 2>/dev/null | tr ' ' '\n' || true)"
  fi
  while IFS= read -r candidate; do
    if valid_ipv4 "$candidate" && private_ipv4 "$candidate"; then
      printf '%s\n' "$candidate"
      return
    fi
  done <<< "$candidates"
  return 1
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --host)
      shift
      [ "$#" -gt 0 ] || fail "--host needs an IPv4 address"
      host="$1"
      ;;
    --env-file)
      shift
      [ "$#" -gt 0 ] || fail "--env-file needs a path"
      env_file="$1"
      ;;
    --dex-config)
      shift
      [ "$#" -gt 0 ] || fail "--dex-config needs a path"
      dex_config="$1"
      dex_config_value="$1"
      ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown option: $1" ;;
  esac
  shift
done

if [ -z "$host" ]; then
  host="$(detect_host)" || fail "could not detect a LAN IPv4 address; rerun with --host ADDRESS"
fi
valid_ipv4 "$host" || fail "--host must be a LAN IPv4 address"
[ ! -e "$env_file" ] || fail "$env_file already exists"
[ ! -e "$dex_config" ] || fail "$dex_config already exists"
[ -f "$repo_root/.env.example" ] || fail ".env.example is missing"
[ -f "$repo_root/deploy/selfhost/dex/config.yaml" ] || fail "Dex template is missing"

mkdir -p "$(dirname "$env_file")" "$(dirname "$dex_config")"
env_tmp="$env_file.tmp.$$"
dex_tmp="$dex_config.tmp.$$"
cleanup() {
  rm -f "$env_tmp" "$dex_tmp"
}
trap cleanup EXIT

awk -v host="$host" -v dex_config="$dex_config_value" '
  /^STUFF_STASH_BIND_ADDRESS=/ { print "STUFF_STASH_BIND_ADDRESS=0.0.0.0"; next }
  /^STUFF_STASH_SELFHOST_HOSTNAME=/ { print "STUFF_STASH_SELFHOST_HOSTNAME=" host; next }
  /^DEX_CONFIG_PATH=/ { print "DEX_CONFIG_PATH=" dex_config; next }
  { gsub(/stuffstash\.localhost/, host); print }
' "$repo_root/.env.example" > "$env_tmp"

awk -v host="$host" '{ gsub(/stuffstash\.localhost/, host); print }' \
  "$repo_root/deploy/selfhost/dex/config.yaml" > "$dex_tmp"
chmod 600 "$env_tmp" "$dex_tmp"
mv "$env_tmp" "$env_file"
mv "$dex_tmp" "$dex_config"
trap - EXIT

echo "Configured Stuff Stash for $host."
echo "Open https://$host:8081 after startup."
echo "Anyone who can reach this server can use the example credentials until you replace them."
