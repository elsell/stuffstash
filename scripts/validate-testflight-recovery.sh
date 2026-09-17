#!/usr/bin/env bash
set -euo pipefail
[[ "${GITHUB_REF:-}" == refs/heads/main ]]
[[ "$#" == 2 ]]
release_tag=$1
build_number=$2
[[ "$release_tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]
[[ "$build_number" =~ ^[1-9][0-9]{0,3}\.[1-9][0-9]?$ ]]
tag_commit=$(git rev-parse --verify "refs/tags/$release_tag^{commit}")
git merge-base --is-ancestor "$tag_commit" refs/remotes/origin/main
