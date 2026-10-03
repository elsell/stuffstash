#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
version="${1:?release tag is required}"
commit="${2:?full release commit is required}"
output="${3:?output directory is required}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'expected stable vMAJOR.MINOR.PATCH release tag' >&2; exit 1; }
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo 'expected full release commit' >&2; exit 1; }
[[ "$(git -C "$root" rev-parse HEAD)" == "$commit" ]] || { echo 'release commit does not match checkout' >&2; exit 1; }
if [[ -n "$(git -C "$root" status --porcelain -- apps/cli scripts/build-cli-release.sh)" ]]; then
  echo 'CLI release inputs must be committed and clean' >&2
  exit 1
fi
mkdir -p "$output"
output="$(cd "$output" && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
export GOWORK=off CGO_ENABLED=0
(cd "$root/apps/cli" && go mod download all)
for os in linux darwin windows; do
  (cd "$root/apps/cli" && GOOS="$os" GOARCH=amd64 go list -buildvcs=false -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/stuffstash)
done | sort -u | sed '/^$/d' > "$tmp/module-paths"
mapfile -t modules < "$tmp/module-paths"
(cd "$root/apps/cli" && go list -m -json "${modules[@]}") > "$tmp/modules.json"
python3 - "$tmp/modules.json" "$tmp/THIRD_PARTY_LICENSES.txt" <<'PYLICENSE'
from pathlib import Path
import json, sys
raw=Path(sys.argv[1]).read_text()
decoder=json.JSONDecoder()
modules=[]
while raw.strip():
    value, end=decoder.raw_decode(raw.lstrip())
    modules.append(value)
    raw=raw.lstrip()[end:]
notices=[]
for module in modules:
    if module.get('Main'): continue
    directory=Path(module['Dir'])
    licenses=sorted(file for file in directory.iterdir() if file.is_file() and file.name.lower().startswith(('license','copying','notice')))
    if not licenses: raise SystemExit('No license found for '+module['Path'])
    notices.append(module['Path']+' '+module['Version'])
    notices.extend(file.name+'\n'+file.read_text(errors='replace') for file in licenses)
Path(sys.argv[2]).write_text('\n\n'.join(notices)+'\n')
PYLICENSE
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
  os="${target%/*}"
  arch="${target#*/}"
  binary="stuffstash"
  if [[ "$os" == windows ]]; then binary=stuffstash.exe; fi
  directory="$tmp/$os-$arch"
  mkdir -p "$directory"
  (cd "$root/apps/cli" && GOOS="$os" GOARCH="$arch" go build -buildvcs=false -trimpath \
    -ldflags="-s -w -X github.com/stuffstash/stuff-stash/cli/internal/version.Tag=$version -X github.com/stuffstash/stuff-stash/cli/internal/version.Commit=$commit" \
    -o "$directory/$binary" ./cmd/stuffstash)
  go version -m "$directory/$binary" > "$directory/build-info.txt"
  python3 - "$directory/build-info.txt" "$version" "$commit" <<'PYVERSION'
from pathlib import Path
import sys
text=Path(sys.argv[1]).read_text()
for field,value in [('Tag',sys.argv[2]),('Commit',sys.argv[3])]:
    if f'github.com/stuffstash/stuff-stash/cli/internal/version.{field}={value}' not in text:
        raise SystemExit('release binary missing expected '+field)
PYVERSION
  if [[ "$os/$arch" == "$(go env GOOS)/$(go env GOARCH)" ]]; then
    "$directory/$binary" version --json > "$directory/version.json"
    python3 - "$directory/version.json" "$version" "$commit" <<'PYSMOKE'
import json,sys
with open(sys.argv[1]) as handle: value=json.load(handle)
if value['version'] != sys.argv[2] or value['commit'] != sys.argv[3]:
    raise SystemExit('CLI reports incorrect release metadata')
PYSMOKE
  fi
  if [[ -f "$root/LICENSE" ]]; then cp "$root/LICENSE" "$directory/LICENSE"; fi
  cp "$tmp/THIRD_PARTY_LICENSES.txt" "$directory/THIRD_PARTY_LICENSES.txt"
  cp "$root/apps/cli/THIRD_PARTY_NOTICES.md" "$directory/THIRD_PARTY_NOTICES.md"
  python3 - "$directory" "$output" "$version" "$os" "$arch" <<'PY'
from pathlib import Path
import gzip, hashlib, io, json, sys, tarfile, zipfile
source, output, version, system, arch = sys.argv[1:]
name = f'stuffstash_{version}_{system}_{arch}'
files = sorted(file for file in Path(source).iterdir() if file.name not in {'build-info.txt', 'version.json'})
if system == 'windows':
    artifact = Path(output) / (name + '.zip')
    with zipfile.ZipFile(artifact, 'w', zipfile.ZIP_DEFLATED) as archive:
        for file in files:
            archive.writestr(zipfile.ZipInfo(file.name, (1980, 1, 1, 0, 0, 0)), file.read_bytes())
else:
    artifact = Path(output) / (name + '.tar.gz')
    with artifact.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', mtime=0, filename='') as compressed, tarfile.open(fileobj=compressed, mode='w') as archive:
        for file in files:
            data = file.read_bytes()
            info = tarfile.TarInfo(file.name)
            info.size = len(data)
            info.mode = 0o755 if file.name == 'stuffstash' else 0o644
            archive.addfile(info, io.BytesIO(data))
checksum = hashlib.sha256(artifact.read_bytes()).hexdigest()
artifact.with_name(artifact.name + '.sha256').write_text(f'{checksum}  {artifact.name}\n')
PY
done
python3 - "$output" "$version" "$commit" <<'PY'
from pathlib import Path
import json, sys
out, version, commit = sys.argv[1:]
assets=[]
for system,arch in [('linux','amd64'),('linux','arm64'),('darwin','amd64'),('darwin','arm64'),('windows','amd64')]:
    filename=f'stuffstash_{version}_{system}_{arch}'+('.zip' if system=='windows' else '.tar.gz')
    checksum=(Path(out)/(filename+'.sha256')).read_text().split()[0]
    assets.append(dict(os=system,architecture=arch,name=filename,sha256=checksum,usbPrinting=False))
(Path(out)/'stuffstash-cli-release.json').write_text(json.dumps(dict(version=version,commit=commit,assets=assets),indent=2)+'\n')
PY
