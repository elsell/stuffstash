#!/usr/bin/env bash
set -euo pipefail

# Run after a scoped Pages commit. Never discard another publisher's changes.
result="$(mktemp)"
trap 'rm -f "$result"' EXIT
for attempt in 1 2 3 4 5; do
  if LC_ALL=C git push --porcelain origin HEAD:refs/heads/gh-pages >"$result" 2>&1; then
    cat "$result"
    exit 0
  fi
  cat "$result" >&2
  if ! LC_ALL=C grep -Eq $'^!\t[^\t]+\t\\[rejected\\] \\((fetch first|non-fast-forward)\\)$' "$result"; then
    exit 1
  fi
  if [ "$attempt" -eq 5 ]; then
    echo 'Pages publication still has concurrent changes after five attempts.' >&2
    exit 1
  fi
  git fetch --no-tags origin refs/heads/gh-pages
  latest="$(git rev-parse FETCH_HEAD)"
  if git merge-base HEAD "$latest" >/dev/null; then
    if ! git rebase "$latest"; then
      git rebase --abort
      echo 'Pages content conflicts with another publication; no changes were pushed.' >&2
      exit 1
    fi
  else
    if ! git rebase --onto "$latest" --root; then
      git rebase --abort
      echo 'Pages initialization conflicts with another publication; no changes were pushed.' >&2
      exit 1
    fi
  fi
done
