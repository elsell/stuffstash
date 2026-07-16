#!/usr/bin/env bash
set -euo pipefail

script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/sign-release-images.sh"
workdir="$(mktemp -d)"

cleanup() {
  rm -rf "$workdir"
}
trap cleanup EXIT

write_fake_cosign() {
  local mode="$1"
  cat >"$workdir/cosign" <<EOF
#!/usr/bin/env bash
set -euo pipefail
count_file="$workdir/count"
count=0
if [ -f "\$count_file" ]; then
  count="\$(cat "\$count_file")"
fi
count=\$((count + 1))
printf '%s' "\$count" >"\$count_file"

case "$mode" in
  transient-once)
    if [ "\$count" -eq 1 ]; then
      echo "Error: signing image: retrieving ID token: reading ID token: fetching ambient OIDC credentials: invalid character 'u' looking for beginning of value" >&2
      exit 1
    fi
    echo "signed \$*"
    ;;
  permanent)
    echo "Error: signing image: registry denied write" >&2
    exit 1
    ;;
  transient-always)
    echo "Error: signing image: retrieving ID token: reading ID token: fetching ambient OIDC credentials: invalid character 'u' looking for beginning of value" >&2
    exit 1
    ;;
esac
EOF
  chmod +x "$workdir/cosign"
  rm -f "$workdir/count"
}

assert_count() {
  local expected="$1"
  local actual
  actual="$(cat "$workdir/count")"
  if [ "$actual" != "$expected" ]; then
    echo "expected cosign to run $expected time(s), got $actual" >&2
    exit 1
  fi
}

write_fake_cosign transient-once
PATH="$workdir:$PATH" RELEASE_COSIGN_SIGN_ATTEMPTS=3 RELEASE_COSIGN_SIGN_RETRY_DELAY_SECONDS=0 "$script_path" "ghcr.io/example/app@sha256:abc" >/dev/null
assert_count 2

write_fake_cosign permanent
if PATH="$workdir:$PATH" RELEASE_COSIGN_SIGN_ATTEMPTS=3 RELEASE_COSIGN_SIGN_RETRY_DELAY_SECONDS=0 "$script_path" "ghcr.io/example/app@sha256:abc" >/dev/null 2>&1; then
  echo "expected permanent signing failure to fail" >&2
  exit 1
fi
assert_count 1

write_fake_cosign transient-always
if PATH="$workdir:$PATH" RELEASE_COSIGN_SIGN_ATTEMPTS=2 RELEASE_COSIGN_SIGN_RETRY_DELAY_SECONDS=0 "$script_path" "ghcr.io/example/app@sha256:abc" >/dev/null 2>&1; then
  echo "expected repeated transient signing failures to fail" >&2
  exit 1
fi
assert_count 2

echo "release image signing retry tests passed"
