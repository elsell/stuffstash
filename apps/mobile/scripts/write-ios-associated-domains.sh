#!/usr/bin/env sh
set -eu

output_path="${1:?provide the generated entitlements output path}"
origin="${EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN:-}"
required="${STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS:-false}"
allow_insecure_local_http="${EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP:-false}"

is_private_local_host() {
  candidate="$1"
  { [ "$candidate" = localhost ] || [ "$candidate" = '[::1]' ]; } && return 0
  case "$candidate" in *[!0-9.]*|'') return 1 ;; esac
  old_ifs=$IFS
  IFS=.
  set -- $candidate
  IFS=$old_ifs
  [ "$#" -eq 4 ] || return 1
  for octet in "$@"; do
    [ -n "$octet" ] && [ "$octet" -le 255 ] 2>/dev/null || return 1
  done
  [ "$1" -eq 127 ] ||
    [ "$1" -eq 10 ] ||
    { [ "$1" -eq 172 ] && [ "$2" -ge 16 ] && [ "$2" -le 31 ]; } ||
    { [ "$1" -eq 192 ] && [ "$2" -eq 168 ]; }
}

case "${required}" in
  1|true|TRUE|yes|YES) required=true ;;
  0|false|FALSE|no|NO|'') required=false ;;
  *) echo 'STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS must be a boolean.' >&2; exit 1 ;;
esac
case "${allow_insecure_local_http}" in
  1|true|TRUE|yes|YES) allow_insecure_local_http=true ;;
  0|false|FALSE|no|NO|'') allow_insecure_local_http=false ;;
  *) echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP must be a boolean.' >&2; exit 1 ;;
esac
if [ "${EAS_BUILD_PROFILE:-}" = production ]; then
  required=true
fi

if [ -z "$origin" ]; then
  if [ "$required" = true ]; then
    echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN is required for production mobile builds.' >&2
    exit 1
  fi
  domain=''
else
  case "$origin" in
    https://*)
      domain="${origin#https://}"
      case "$domain" in
        ''|*/*|*:*|*'?'*|*'#'*|*'@'*)
          echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.' >&2
          exit 1
          ;;
      esac
      case "$domain" in
        *[!A-Za-z0-9.-]*|.*|*.|*..*)
          echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.' >&2
          exit 1
          ;;
      esac
      ;;
    http://*)
      if [ "$required" = true ] || [ "$allow_insecure_local_http" != true ]; then
        echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.' >&2
        exit 1
      fi
      authority="${origin#http://}"
      case "$authority" in
        ''|*/*|*'?'*|*'#'*|*'@'*)
          echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a private local HTTP origin.' >&2
          exit 1
          ;;
      esac
      if [ "$authority" = '[::1]' ]; then
        host='[::1]'
      elif [ "${authority#'[::1]:'}" != "$authority" ]; then
        host='[::1]'
        port="${authority#'[::1]:'}"
        case "$port" in ''|*[!0-9]*) echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a private local HTTP origin.' >&2; exit 1 ;; esac
      else
        case "$authority" in *:*:*) echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a private local HTTP origin.' >&2; exit 1 ;; esac
        host="${authority%%:*}"
      fi
      if [ "$host" != "$authority" ] && [ "$host" != '[::1]' ]; then
        port="${authority##*:}"
        case "$port" in ''|*[!0-9]*) echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a private local HTTP origin.' >&2; exit 1 ;; esac
      fi
      if ! is_private_local_host "$host"; then
        echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a private local HTTP origin.' >&2
        exit 1
      fi
      domain=''
      ;;
    *) echo 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.' >&2; exit 1 ;;
  esac
fi

mkdir -p "$(dirname "$output_path")"
if [ -n "$domain" ]; then
  domains="
    <key>com.apple.developer.associated-domains</key>
    <array>
      <string>applinks:$domain</string>
    </array>"
else
  domains=''
fi
cat > "$output_path" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>$domains
  </dict>
</plist>
EOF
