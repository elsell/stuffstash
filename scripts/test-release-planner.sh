#!/usr/bin/env bash
set -euo pipefail

script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/plan-release.sh"
workdir="$(mktemp -d)"

cleanup() {
  rm -rf "$workdir"
}
trap cleanup EXIT

assert_output_contains() {
  local output="$1"
  local expected="$2"
  if ! printf '%s\n' "$output" | grep -Fxq "$expected"; then
    echo "expected release planner output to contain: $expected" >&2
    echo "actual output:" >&2
    printf '%s\n' "$output" >&2
    exit 1
  fi
}

commit() {
  local message="$1"
  git commit --allow-empty -m "$message" >/dev/null
}

git -C "$workdir" init >/dev/null
git -C "$workdir" config user.name "Stuff Stash Test"
git -C "$workdir" config user.email "stuffstash@example.test"
cd "$workdir"

commit "chore: initial repository"
output="$("$script_path")"
assert_output_contains "$output" "release_required=false"
assert_output_contains "$output" "bump=none"

commit "feat: add inventory search"
output="$("$script_path")"
assert_output_contains "$output" "release_required=true"
assert_output_contains "$output" "next_tag=v0.1.0"
assert_output_contains "$output" "bump=minor"

git tag v0.1.0
commit "fix: preserve attachment metadata"
output="$("$script_path")"
assert_output_contains "$output" "previous_tag=v0.1.0"
assert_output_contains "$output" "next_tag=v0.1.1"
assert_output_contains "$output" "bump=patch"

commit "feat!: change authorization model"
output="$("$script_path")"
assert_output_contains "$output" "next_tag=v1.0.0"
assert_output_contains "$output" "bump=major"

git tag v1.0.0
commit "docs: clarify local setup"
output="$("$script_path")"
assert_output_contains "$output" "release_required=false"
assert_output_contains "$output" "next_tag=v1.0.0"
assert_output_contains "$output" "bump=none"

echo "release planner tests passed"
