#!/usr/bin/env python3
"""Offline install documentation from release-verified, checked-in metadata."""
import argparse
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
METADATA = ROOT / 'docs/src/data/cli-release.json'
DOCUMENT = ROOT / 'docs/src/content/docs/cli-downloads.md'


def render(value):
    text = '---\ntitle: Download the CLI\ndescription: Versioned Stuff Stash command-line downloads.\n---\n\n'
    if value is None:
        return text + 'Published CLI downloads are not available yet. [Build from source](../cli/).\n'
    tag = value['version']
    if not re.fullmatch(r'v\d+\.\d+\.\d+', tag) or not re.fullmatch('[0-9a-f]{40}', value['commit']):
        raise ValueError('Invalid release identity')
    repo = value['repository']
    if not re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', repo) or not value['publishedAt']:
        raise ValueError('Invalid published release')
    text += f"Version **{tag}**. Download the archive for your computer and verify its checksum before extracting it.\n\n"
    text += 'USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.\n\n'
    names = {'linux': 'Linux', 'darwin': 'macOS', 'windows': 'Windows'}
    seen = set()
    for asset in value['assets']:
        system, arch = asset['os'], asset['architecture']
        if (system, arch) not in {('linux','amd64'), ('linux','arm64'), ('darwin','amd64'), ('darwin','arm64'), ('windows','amd64')} or (system, arch) in seen:
            raise ValueError('Invalid platform asset')
        seen.add((system, arch))
        name = f'stuffstash_{tag}_{system}_{arch}' + ('.zip' if system == 'windows' else '.tar.gz')
        sha = asset['sha256']
        if asset['name'] != name or not re.fullmatch('[0-9a-f]{64}', sha):
            raise ValueError('Invalid archive identity')
        url = f'https://github.com/{repo}/releases/download/{tag}/{name}'
        text += f'## {names[system]} {arch}\n\n'
        if system == 'windows':
            text += f'```powershell\ncurl.exe --fail --location --output {name} {url}\n(Get-FileHash {name} -Algorithm SHA256).Hash.ToLower()\n```\n\nExpected SHA-256: `{sha}`. Extract only after it matches.\n\n'
        else:
            verify = 'sha256sum --check' if system == 'linux' else 'shasum -a 256 --check'
            text += f"```sh\ncurl --fail --location --output {name} {url}\nprintf '%s  %s\\n' '{sha}' '{name}' | {verify}\n```\n\nAfter verification succeeds:\n\n```sh\ntar -xzf {name}\n./stuffstash version\n```\n\n"
    text += f"[Release notes and all assets](https://github.com/{repo}/releases/tag/{tag}). Source commit: `{value['commit']}`.\n\nContinue with [sign-in and inventory commands](../cli/#sign-in).\n"
    return text


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--update', type=Path)
    args = parser.parse_args()
    value = json.loads(METADATA.read_text())
    if args.update:
        candidate = json.loads(args.update.read_text())
        render(candidate)
        version = lambda data: tuple(map(int, data['version'][1:].split('.')))
        if value is not None and version(candidate) < version(value):
            return
        if value is not None and candidate['version'] == value['version'] and candidate != value:
            raise ValueError('Published release metadata changed for an existing version')
        value = candidate
        METADATA.write_text(json.dumps(value, indent=2) + '\n')
    expected = render(value)
    if args.check:
        if not DOCUMENT.exists() or DOCUMENT.read_text() != expected:
            raise SystemExit('CLI download docs drift: run python3 scripts/generate-cli-downloads.py')
    else:
        DOCUMENT.write_text(expected)


if __name__ == '__main__':
    main()
