#!/usr/bin/env bash
set -euo pipefail

attempts="${RELEASE_COSIGN_SIGN_ATTEMPTS:-4}"
delay_seconds="${RELEASE_COSIGN_SIGN_RETRY_DELAY_SECONDS:-10}"

if ! [[ "$attempts" =~ ^[1-9][0-9]*$ ]]; then
  echo "RELEASE_COSIGN_SIGN_ATTEMPTS must be a positive integer." >&2
  exit 2
fi

if ! [[ "$delay_seconds" =~ ^[0-9]+$ ]]; then
  echo "RELEASE_COSIGN_SIGN_RETRY_DELAY_SECONDS must be a non-negative integer." >&2
  exit 2
fi

if [ "$#" -eq 0 ]; then
  echo "usage: scripts/sign-release-images.sh <image-ref> [<image-ref>...]" >&2
  exit 2
fi

is_transient_oidc_error() {
  local log_file="$1"
  grep -Eq "fetching ambient OIDC credentials|retrieving ID token|reading ID token" "$log_file"
}

sign_image() {
  local image_ref="$1"
  local attempt=1
  local log_file
  log_file="$(mktemp)"

  while [ "$attempt" -le "$attempts" ]; do
    if cosign sign --yes "$image_ref" >"$log_file" 2>&1; then
      cat "$log_file"
      rm -f "$log_file"
      return 0
    else
      local status=$?
    fi

    cat "$log_file" >&2

    if is_transient_oidc_error "$log_file" && [ "$attempt" -lt "$attempts" ]; then
      echo "cosign signing failed while retrieving the GitHub Actions OIDC token; retrying attempt $((attempt + 1)) of $attempts." >&2
      sleep "$delay_seconds"
      attempt=$((attempt + 1))
      : >"$log_file"
      continue
    fi

    rm -f "$log_file"
    return "$status"
  done
}

for image_ref in "$@"; do
  if [ -z "$image_ref" ]; then
    echo "image reference must not be empty." >&2
    exit 2
  fi
  sign_image "$image_ref"
done
