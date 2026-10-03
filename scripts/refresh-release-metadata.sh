#!/usr/bin/env bash
set -euo pipefail
publication="$(realpath "${1:?publication directory}")"
verified_metadata="$(realpath "${2:?verified metadata file}")"
if [[ "$(cat "$verified_metadata")" == null ]]; then
  echo 'Repaired release is older than a published stable release; leaving current download metadata unchanged.'
  exit 0
fi
tag="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["tag"])' "$publication/publication.json")"
api_image="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["apiImage"])' "$publication/publication.json")"
web_image="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["webImage"])' "$publication/publication.json")"
# Start maintenance from current main, never rewind unrelated changes to an old repair commit.
git fetch origin main
git checkout -B "release/metadata-$tag" origin/main
python3 scripts/generate-cli-downloads.py --update "$verified_metadata"
# An old repair must not roll self-host defaults back after newer CLI publication.
current="$(python3 -c 'import json; print(json.load(open("docs/src/data/cli-release.json"))["version"])')"
if [[ "$current" == "$tag" ]]; then
  node scripts/update-selfhost-image-refs.mjs --api-image "$api_image" --web-image "$web_image"
fi
git add .env.example docs/src/data/cli-release.json docs/src/content/docs/cli-downloads.md
if git diff --cached --quiet; then
  gh workflow run docs-pages.yml --ref main
  exit 0
fi
git config user.name 'github-actions[bot]'
git config user.email '41898282+github-actions[bot]@users.noreply.github.com'
git commit -m "chore(release): refresh downloads and self-host images for $tag"
branch="release/metadata-$tag"
if git ls-remote --exit-code --heads origin "$branch" >/dev/null; then
  git fetch origin "refs/heads/$branch:refs/remotes/origin/$branch"
  git push --force-with-lease="refs/heads/$branch:$(git rev-parse "refs/remotes/origin/$branch")" origin "HEAD:$branch"
else
  git push origin "HEAD:$branch"
fi
body="$(mktemp)"
trap 'rm -f "$body"' EXIT
printf 'Updates CLI download links and self-host image digests from verified published release `%s`. The generated download commands pin the release and checksum.\n' "$tag" > "$body"
number="$(gh pr list --head "$branch" --base main --state open --json number --jq '.[0].number // ""')"
if [[ -z "$number" ]]; then
  gh pr create --base main --head "$branch" --title "chore(release): refresh downloads and self-host images for $tag" --body-file "$body"
  number="$(gh pr list --head "$branch" --base main --state open --json number --jq '.[0].number')"
fi
gh workflow run ci.yml --ref "$branch"
gh pr merge "$number" --auto --squash --delete-branch
# GITHUB_TOKEN merges do not emit downstream push workflows. Dispatch only once merged.
for ((attempt=0; attempt<90; attempt++)); do
  state="$(gh pr view "$number" --json state --jq .state)"
  if [[ "$state" == MERGED ]]; then
    gh workflow run docs-pages.yml --ref main
    exit 0
  fi
  if [[ "$state" == CLOSED ]]; then echo 'Release metadata PR closed without merging' >&2; exit 1; fi
  sleep 20
done
echo "Metadata PR #$number has not merged yet; retry publication repair or dispatch Docs Pages after merging." >&2
exit 1
