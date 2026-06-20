#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/plan-release.sh [--github-output]

Plans the next SemVer release from Conventional Commit messages since the
latest vMAJOR.MINOR.PATCH tag. The script prints shell assignments:

  release_required=true|false
  previous_tag=vX.Y.Z|none
  next_version=X.Y.Z
  next_tag=vX.Y.Z
  bump=major|minor|patch|none

Use --github-output inside GitHub Actions to also append those values to
$GITHUB_OUTPUT.
USAGE
}

write_github_output=false
if [ "${1:-}" = "--github-output" ]; then
  write_github_output=true
elif [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
  usage
  exit 0
elif [ "$#" -gt 0 ]; then
  usage >&2
  exit 2
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "plan-release must run inside a git repository" >&2
  exit 1
fi

latest_tag="$(git tag --merged HEAD --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [ -n "$latest_tag" ]; then
  range="${latest_tag}..HEAD"
  previous_tag="$latest_tag"
  previous_version="${latest_tag#v}"
else
  range="HEAD"
  previous_tag="none"
  previous_version="0.0.0"
fi

commit_count="$(git rev-list --count "$range")"
if [ "$commit_count" = "0" ]; then
  bump="none"
else
  subjects="$(git log --format='%s' "$range")"
  bodies="$(git log --format='%b' "$range")"

  bump="none"
  if printf '%s\n%s\n' "$subjects" "$bodies" | grep -Eq '(^|[[:space:]])BREAKING CHANGE:'; then
    bump="major"
  elif printf '%s\n' "$subjects" | grep -Eq '^[a-zA-Z]+(\([^)]+\))?!:'; then
    bump="major"
  elif printf '%s\n' "$subjects" | grep -Eq '^feat(\([^)]+\))?:'; then
    bump="minor"
  elif printf '%s\n' "$subjects" | grep -Eq '^(fix|perf)(\([^)]+\))?:'; then
    bump="patch"
  fi
fi

IFS=. read -r major minor patch <<EOF
$previous_version
EOF

case "$bump" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
  none)
    ;;
  *)
    echo "unsupported release bump: $bump" >&2
    exit 1
    ;;
esac

next_version="${major}.${minor}.${patch}"
next_tag="v${next_version}"
release_required=false
if [ "$bump" != "none" ]; then
  release_required=true
fi

emit() {
  local key="$1"
  local value="$2"
  printf '%s=%s\n' "$key" "$value"
  if [ "$write_github_output" = true ]; then
    if [ -z "${GITHUB_OUTPUT:-}" ]; then
      echo "--github-output requires GITHUB_OUTPUT" >&2
      exit 1
    fi
    printf '%s=%s\n' "$key" "$value" >> "$GITHUB_OUTPUT"
  fi
}

emit release_required "$release_required"
emit previous_tag "$previous_tag"
emit next_version "$next_version"
emit next_tag "$next_tag"
emit bump "$bump"
