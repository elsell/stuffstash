#!/usr/bin/env bash
set -euo pipefail

container_name="stuff-stash-garage-test-$$"
bucket="stuffstash-test"
access_key="GK0123456789abcdef0123456789abcdef"
secret_key="0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
workdir="$(mktemp -d)"
go_binary="${GO:-go}"

cleanup() {
  docker rm -f "$container_name" >/dev/null 2>&1 || true
  rm -rf "$workdir"
}
trap cleanup EXIT

garage_image_for_platform() {
  case "$(uname -m)" in
    arm64|aarch64)
      printf '%s\n' "dxflrs/garage:v2.3.0@sha256:2d3f94a89a8a02dc49fa75594d6df67ed9c6ffe08fe55ed023d0c9776f71a9bd"
      ;;
    x86_64|amd64)
      printf '%s\n' "dxflrs/garage:v2.3.0@sha256:dac0c92add4f1a0b41035e94b41036a270ffbe88a37c7ac9c3f19e6dc5bdccf2"
      ;;
    *)
      echo "unsupported architecture for pinned Garage verifier: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

garage_image="${GARAGE_IMAGE:-$(garage_image_for_platform)}"
case "$garage_image" in
  *@sha256:*) ;;
  *)
    echo "GARAGE_IMAGE must be pinned with @sha256:" >&2
    exit 1
    ;;
esac

cat > "$workdir/garage.toml" <<EOF
metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"

replication_factor = 1
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:3900"
root_domain = ".s3.garage.localhost"

[s3_web]
bind_addr = "[::]:3902"
root_domain = ".web.garage.localhost"
index = "index.html"

[admin]
api_bind_addr = "[::]:3903"
admin_token = "local-garage-admin-token"
metrics_token = "local-garage-metrics-token"
EOF

mkdir -p "$workdir/meta" "$workdir/data"

echo "starting Garage ${garage_image}"
docker run --rm -d \
  --name "$container_name" \
  -p 127.0.0.1::3900 \
  -v "$workdir/garage.toml:/etc/garage.toml:ro" \
  -v "$workdir/meta:/var/lib/garage/meta" \
  -v "$workdir/data:/var/lib/garage/data" \
  -e "GARAGE_DEFAULT_ACCESS_KEY=$access_key" \
  -e "GARAGE_DEFAULT_SECRET_KEY=$secret_key" \
  -e "GARAGE_DEFAULT_BUCKET=$bucket" \
  "$garage_image" \
  /garage server --single-node --default-bucket >/dev/null

host_port=""
for _ in $(seq 1 120); do
  if docker exec "$container_name" /garage status >/dev/null 2>&1; then
    port_mapping="$(docker port "$container_name" 3900/tcp | head -n 1)"
    host_port="${port_mapping##*:}"
    if [ -n "$host_port" ]; then
      break
    fi
  fi
  sleep 0.5
done

if [ -z "$host_port" ]; then
  docker logs "$container_name" >&2 || true
  echo "Garage did not become ready" >&2
  exit 1
fi

echo "verifying S3 blob adapter against Garage on 127.0.0.1:${host_port}"
STUFF_STASH_TEST_S3_ENDPOINT="127.0.0.1:${host_port}" \
STUFF_STASH_TEST_S3_ACCESS_KEY="$access_key" \
STUFF_STASH_TEST_S3_SECRET_KEY="$secret_key" \
STUFF_STASH_TEST_S3_BUCKET="$bucket" \
GOCACHE="${GOCACHE:-$PWD/.cache/go-build}" \
"$go_binary" test ./apps/api/internal/adapters/blobstore -run TestS3StoreAgainstGarage -count=1 -timeout=2m
