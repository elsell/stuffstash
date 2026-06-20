#!/usr/bin/env python3
import argparse
import concurrent.futures
import datetime as dt
import json
import os
import re
import shutil
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path


UTC = dt.timezone.utc
PSEUDO_VERSION_RE = re.compile(r"-(\d{14})-[0-9a-f]{12,}(?:\.\d+)?$")
PNPM_PACKAGE_RE = re.compile(r"^  '?(?P<key>[^':]+@[^':]+)'?:$")


def parse_time(value):
    if not value or value == "0001-01-01T00:00:00Z":
        return None
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    return dt.datetime.fromisoformat(value).astimezone(UTC)


def fetch_json(url):
    if shutil.which("curl"):
        result = subprocess.run(
            ["curl", "-fsSL", "--proto", "=https", "--tlsv1.2", url],
            check=True,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        return json.loads(result.stdout)
    request = urllib.request.Request(url, headers={"Accept": "application/json"})
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.loads(response.read().decode("utf-8"))


def npm_package_versions(lockfile):
    in_packages = False
    versions = set()
    for line in lockfile.read_text(encoding="utf-8").splitlines():
        if line == "packages:":
            in_packages = True
            continue
        if in_packages and line and not line.startswith(" "):
            break
        if not in_packages:
            continue
        match = PNPM_PACKAGE_RE.match(line)
        if not match:
            continue
        key = match.group("key")
        name, version = key.rsplit("@", 1)
        if not name or not version:
            continue
        version = version.split("(", 1)[0]
        if version.startswith(("link:", "workspace:", "file:")):
            continue
        versions.add((name, version, str(lockfile)))
    return versions


def npm_published_at(name, version):
    escaped_name = urllib.parse.quote(name, safe="")
    data = fetch_json(f"https://registry.npmjs.org/{escaped_name}")
    try:
        return parse_time(data["time"][version])
    except KeyError as exc:
        raise RuntimeError(f"npm metadata for {name}@{version} did not include a publish time") from exc


def go_modules(go_mod):
    command = ["go", "list", "-m", "-json", "all"]
    env = os.environ.copy()
    env["GOWORK"] = "off"
    result = subprocess.run(
        command,
        cwd=go_mod.parent,
        env=env,
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    modules = []
    decoder = json.JSONDecoder()
    text = result.stdout.lstrip()
    while text:
        module, index = decoder.raw_decode(text)
        text = text[index:].lstrip()
        if module.get("Main") or "Version" not in module:
            continue
        modules.append((module["Path"], module["Version"], module.get("Time"), str(go_mod)))
    return modules


def pseudo_version_time(version):
    match = PSEUDO_VERSION_RE.search(version)
    if not match:
        return None
    return dt.datetime.strptime(match.group(1), "%Y%m%d%H%M%S").replace(tzinfo=UTC)


def go_published_at(path, version, module_time):
    parsed = parse_time(module_time)
    if parsed is not None:
        return parsed
    parsed = pseudo_version_time(version)
    if parsed is not None:
        return parsed
    raise RuntimeError(f"Go metadata for {path}@{version} did not include a publish time")


def check_age(kind, name, version, published_at, source, cutoff):
    if published_at is None:
        return f"{kind} {name}@{version} from {source} has no known publish time"
    if published_at > cutoff:
        return (
            f"{kind} {name}@{version} from {source} was published "
            f"{published_at.isoformat()} after cutoff {cutoff.isoformat()}"
        )
    return None


def main():
    parser = argparse.ArgumentParser(description="Reject npm and Go dependencies newer than the allowed age.")
    parser.add_argument("--min-age-days", type=int, default=14)
    parser.add_argument("--now", help="UTC timestamp override for tests, for example 2026-06-20T00:00:00Z")
    parser.add_argument("--skip-npm", action="store_true")
    parser.add_argument("--skip-go", action="store_true")
    args = parser.parse_args()

    root = Path.cwd()
    now = parse_time(args.now) if args.now else dt.datetime.now(UTC)
    cutoff = now - dt.timedelta(days=args.min_age_days)
    failures = []

    if not args.skip_npm:
        lockfiles = sorted(root.glob("**/pnpm-lock.yaml"))
        lockfiles = [path for path in lockfiles if "node_modules" not in path.parts]
        npm_versions = set()
        for lockfile in lockfiles:
            npm_versions.update(npm_package_versions(lockfile))
        def check_npm(item):
            name, version, source = item
            try:
                published_at = npm_published_at(name, version)
                return check_age("npm", name, version, published_at, source, cutoff)
            except (RuntimeError, urllib.error.URLError, TimeoutError, subprocess.CalledProcessError) as exc:
                return f"npm {name}@{version} from {source} metadata check failed: {exc}"

        with concurrent.futures.ThreadPoolExecutor(max_workers=16) as executor:
            for failure in executor.map(check_npm, sorted(npm_versions)):
                if failure:
                    failures.append(failure)

    if not args.skip_go:
        go_mods = sorted(root.glob("**/go.mod"))
        go_mods = [path for path in go_mods if "node_modules" not in path.parts]
        for go_mod in go_mods:
            try:
                modules = go_modules(go_mod)
            except subprocess.CalledProcessError as exc:
                failures.append(f"go module graph failed for {go_mod}: {exc.stderr.strip()}")
                continue
            for path, version, module_time, source in modules:
                try:
                    published_at = go_published_at(path, version, module_time)
                    failure = check_age("go", path, version, published_at, source, cutoff)
                except RuntimeError as exc:
                    failure = str(exc)
                if failure:
                    failures.append(failure)

    if failures:
        print("dependency age check failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1

    print(f"dependency age check passed; all checked versions were published before {cutoff.isoformat()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
