#!/usr/bin/env sh
set -eu

api_base_url="${STUFF_STASH_WEB_API_BASE_URL:-${STUFF_STASH_API_ORIGIN:-http://localhost:8080}}"
oidc_issuer="${STUFF_STASH_WEB_OIDC_ISSUER:-${STUFF_STASH_OIDC_ISSUER:-http://localhost:5556/dex}}"
oidc_client_id="${STUFF_STASH_WEB_OIDC_CLIENT_ID:-${STUFF_STASH_OIDC_CLIENT_ID:-stuff-stash-web-local}}"
oidc_redirect_uri="${STUFF_STASH_WEB_OIDC_REDIRECT_URI:-${STUFF_STASH_WEB_ORIGIN:-http://localhost:8081}/callback}"
media_max_bytes="${STUFF_STASH_WEB_MEDIA_MAX_BYTES:-${STUFF_STASH_MAX_ATTACHMENT_BYTES:-5242880}}"
s3_endpoint="${STUFF_STASH_S3_PUBLIC_ENDPOINT:-localhost:3900}"
s3_secure="${STUFF_STASH_S3_SECURE:-false}"

require_safe_value() {
  name="$1"
  value="$2"
  if printf '%s' "$value" | grep '[\\|"[:cntrl:]]' >/dev/null 2>&1; then
      echo "$name contains characters that are not supported in web runtime configuration" >&2
      exit 1
  fi
}

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

origin_from_url() {
  value="$1"
  case "$value" in
    http://*)
      rest="${value#http://}"
      printf 'http://%s' "${rest%%/*}"
      ;;
    https://*)
      rest="${value#https://}"
      printf 'https://%s' "${rest%%/*}"
      ;;
    *)
      printf '%s' "$value"
      ;;
  esac
}

s3_origin_from_endpoint() {
  endpoint="$1"
  case "$endpoint" in
    http://*|https://*)
      origin_from_url "$endpoint"
      ;;
    *)
      if [ "$s3_secure" = "true" ] || [ "$s3_secure" = "1" ] || [ "$s3_secure" = "yes" ]; then
        printf 'https://%s' "$endpoint"
      else
        printf 'http://%s' "$endpoint"
      fi
      ;;
  esac
}

require_safe_value STUFF_STASH_WEB_API_BASE_URL "$api_base_url"
require_safe_value STUFF_STASH_WEB_OIDC_ISSUER "$oidc_issuer"
require_safe_value STUFF_STASH_WEB_OIDC_CLIENT_ID "$oidc_client_id"
require_safe_value STUFF_STASH_WEB_OIDC_REDIRECT_URI "$oidc_redirect_uri"
require_safe_value STUFF_STASH_S3_PUBLIC_ENDPOINT "$s3_endpoint"

case "$media_max_bytes" in
  ''|*[!0-9]*)
    echo "STUFF_STASH_WEB_MEDIA_MAX_BYTES must be a positive integer" >&2
    exit 1
    ;;
  0)
    echo "STUFF_STASH_WEB_MEDIA_MAX_BYTES must be greater than zero" >&2
    exit 1
    ;;
esac

cat > /tmp/stuffstash-config.json <<EOF
{
  "apiBaseUrl": "$(json_escape "$api_base_url")",
  "oidcIssuer": "$(json_escape "$oidc_issuer")",
  "oidcClientId": "$(json_escape "$oidc_client_id")",
  "oidcRedirectUri": "$(json_escape "$oidc_redirect_uri")",
  "mediaUploadPolicy": {
    "supportedContentTypes": ["image/jpeg", "image/png", "image/webp", "application/pdf"],
    "maxBytes": $media_max_bytes
  }
}
EOF

/usr/local/bin/write-mobile-association-files /tmp

nginx_template="/opt/app-root/etc/nginx.default.d/stuffstash.conf"
nginx_server_conf="/tmp/stuffstash-server.conf"
api_origin="$(origin_from_url "$api_base_url")"
oidc_origin="$(origin_from_url "$oidc_issuer")"
s3_origin="$(s3_origin_from_endpoint "$s3_endpoint")"

sed \
  -e "s|__STUFFSTASH_CSP_CONNECT_SRC__|$api_origin $oidc_origin $s3_origin|g" \
  -e "s|__STUFFSTASH_CSP_FORM_ACTION__|$oidc_origin|g" \
  -e "s|__STUFFSTASH_CSP_IMG_SRC__|$api_origin $s3_origin|g" \
  "$nginx_template" > "$nginx_server_conf"

cat > /tmp/nginx.conf <<'EOF'
worker_processes auto;
error_log /dev/stderr warn;
pid /tmp/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    access_log /dev/stdout;
    sendfile on;

    client_body_temp_path /tmp/client_body_temp;
    proxy_temp_path /tmp/proxy_temp;
    fastcgi_temp_path /tmp/fastcgi_temp;
    uwsgi_temp_path /tmp/uwsgi_temp;
    scgi_temp_path /tmp/scgi_temp;

    server {
        listen 8080;
        root /opt/app-root/src;
        include /tmp/stuffstash-server.conf;
    }
}
EOF

exec nginx -e /dev/stderr -c /tmp/nginx.conf -g "daemon off;"
