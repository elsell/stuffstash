#!/usr/bin/env bash
set -euo pipefail
# Hosted macOS only: no workstation compilation or reusable service credentials.
[[ "$(uname -s)" == Darwin && "${GITHUB_ACTIONS:-}" == true ]] || exit 1
root="$(cd "$(dirname "$0")/../.." && pwd)"
work="${RUNNER_TEMP:?}/connected-native"
mkdir -p "$work/bin" "$work/source"
fetch_verified() {
  curl --fail --silent --show-error --location --max-time 120 "$1" -o "$2"
  printf '%s  %s\n' "$3" "$2" | shasum -a 256 -c -
}
case "$(uname -m)" in
  arm64) arch=arm64; digest=ad8422188c24487e233e1a7692bbe060c690c81a91cee2feb54993c968fb639b ;;
  x86_64) arch=amd64; digest=67dc53b60cf71224cfc819451eda0cfd328a3bb0f7f119decec79e53e0e0b3d8 ;;
  *) exit 1 ;;
esac
fetch_verified "https://github.com/authzed/spicedb/releases/download/v1.47.1/spicedb_1.47.1_darwin_${arch}.tar.gz" "$work/spicedb.tar.gz" "$digest"
tar -xzf "$work/spicedb.tar.gz" -C "$work/bin" spicedb
fetch_verified 'https://codeload.github.com/dexidp/dex/tar.gz/2dce75009acfcd9df77d57f2d144aae999a4c569' "$work/dex.tar.gz" 36cfcc2196f1d5188563cb5287c85926b9c752394268c093ff91e7067a269880
tar -xzf "$work/dex.tar.gz" -C "$work/source"
dex_source="$work/source/dex-2dce75009acfcd9df77d57f2d144aae999a4c569"
(cd "$dex_source" && GOWORK=off go build -mod=readonly -o "$work/bin/dex" ./cmd/dex)
(cd "$root/apps/api" && go build -mod=readonly -o "$work/bin/stuff-stash" ./cmd/stuff-stash)
rm "$work/spicedb.tar.gz" "$work/dex.tar.gz"
