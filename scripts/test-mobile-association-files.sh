#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp_directory=$(mktemp -d)
trap 'rm -rf "$tmp_directory"' EXIT HUP INT TERM

fingerprint='AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99'
STUFF_STASH_MOBILE_IOS_APP_ID='7585W4AG8C.app.stuffstash.mobile' \
STUFF_STASH_MOBILE_ANDROID_PACKAGE='app.stuffstash.mobile' \
STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT="$fingerprint" \
  "$repo_root/deploy/web/write-mobile-association-files.sh" "$tmp_directory"

python3 - "$tmp_directory" "$fingerprint" <<'PY'
import json
import pathlib
import sys

directory = pathlib.Path(sys.argv[1])
fingerprint = sys.argv[2]
aasa = json.loads((directory / "apple-app-site-association").read_text())
assert aasa == {
    "applinks": {
        "details": [{
            "appIDs": ["7585W4AG8C.app.stuffstash.mobile"],
            "components": [{"/": "/invitations/accept", "comment": "Stuff Stash inventory invitations"}],
        }]
    }
}
assetlinks = json.loads((directory / "assetlinks.json").read_text())
assert assetlinks[0]["target"]["package_name"] == "app.stuffstash.mobile"
assert assetlinks[0]["target"]["sha256_cert_fingerprints"] == [fingerprint]
PY

STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT='' \
STUFF_STASH_MOBILE_IOS_APP_ID='' \
  "$repo_root/deploy/web/write-mobile-association-files.sh" "$tmp_directory"
test "$(cat "$tmp_directory/assetlinks.json")" = '[]'
test "$(cat "$tmp_directory/apple-app-site-association")" = '{"applinks":{"details":[]}}'

if STUFF_STASH_MOBILE_IOS_APP_ID='not-an-app-id' \
  "$repo_root/deploy/web/write-mobile-association-files.sh" "$tmp_directory" >/dev/null 2>&1; then
  echo 'invalid Apple application identifier was accepted' >&2
  exit 1
fi

if STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT='AA:BB' \
  "$repo_root/deploy/web/write-mobile-association-files.sh" "$tmp_directory" >/dev/null 2>&1; then
  echo 'invalid Android certificate fingerprint was accepted' >&2
  exit 1
fi

python3 - "$repo_root/deploy/web/nginx.conf" "$repo_root/deploy/selfhost/caddy/Caddyfile" <<'PY'
import pathlib
import re
import sys

nginx = pathlib.Path(sys.argv[1]).read_text()
for route, file_name in (
    ("/.well-known/apple-app-site-association", "apple-app-site-association"),
    ("/.well-known/assetlinks.json", "assetlinks.json"),
):
    match = re.search(rf"location = {re.escape(route)}\s*\{{([^}}]*)\}}", nginx, re.DOTALL)
    assert match, f"missing exact nginx route for {route}"
    block = match.group(1)
    assert "default_type application/json;" in block
    assert f"alias /tmp/{file_name};" in block
    assert 'add_header Cache-Control "public, max-age=3600" always;' in block
    assert not re.search(r"\b(return|rewrite|proxy_pass|try_files)\b", block), f"{route} may redirect or fall through"

caddy = pathlib.Path(sys.argv[2]).read_text()
web_host = re.search(r"https://\{\$STUFF_STASH_SELFHOST_HOSTNAME\}:\{\$STUFF_STASH_WEB_PORT\}\s*\{([^}]*)\}", caddy, re.DOTALL)
assert web_host, "missing self-host web HTTPS block"
assert "reverse_proxy web:8080" in web_host.group(1)
assert "redir" not in web_host.group(1)
PY

grep -q '/usr/local/bin/write-mobile-association-files /tmp' "$repo_root/deploy/web/start-web-runtime.sh"
grep -q '/usr/local/bin/write-mobile-association-files' "$repo_root/Dockerfile.web"
