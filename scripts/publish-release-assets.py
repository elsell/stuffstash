#!/usr/bin/env python3
"""Retryable publication of exact, previously built release assets through gh."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

TAG = re.compile(r'v(\d+)\.(\d+)\.(\d+)\Z')
COMMIT = re.compile(r'[0-9a-f]{40}\Z')
NAME = re.compile(r'[A-Za-z0-9][A-Za-z0-9._-]*\Z')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate_plan(root):
    plan = json.loads((root / 'publication.json').read_text())
    if not TAG.fullmatch(plan['tag']) or not COMMIT.fullmatch(plan['commit']):
        raise ValueError('Invalid publication identity')
    names = set()
    for asset in plan['assets']:
        name = asset['name']
        if not NAME.fullmatch(name) or name in names or not re.fullmatch('[0-9a-f]{64}', asset['sha256']):
            raise ValueError('Invalid or duplicated asset')
        names.add(name)
        if digest(root / name) != asset['sha256']:
            raise ValueError('Staged artifact checksum mismatch: ' + name)
    cli = json.loads((root / 'stuffstash-cli-release.json').read_text())
    expected = {('linux', 'amd64'), ('linux', 'arm64'), ('darwin', 'amd64'), ('darwin', 'arm64'), ('windows', 'amd64')}
    if cli['version'] != plan['tag'] or cli['commit'] != plan['commit']:
        raise ValueError('CLI provenance differs from publication')
    if {(item['os'], item['architecture']) for item in cli['assets']} != expected or len(cli['assets']) != 5:
        raise ValueError('Incomplete CLI release matrix')
    for asset in cli['assets']:
        suffix = '.zip' if asset['os'] == 'windows' else '.tar.gz'
        expected_name = f"stuffstash_{plan['tag']}_{asset['os']}_{asset['architecture']}{suffix}"
        if asset['name'] != expected_name or digest(root / expected_name) != asset['sha256']:
            raise ValueError('Invalid CLI archive identity')
        if (root / (expected_name + '.sha256')).read_text() != f"{asset['sha256']}  {expected_name}\n":
            raise ValueError('Invalid CLI checksum file')
        if not {expected_name, expected_name + '.sha256'} <= names:
            raise ValueError('CLI assets absent from publication plan')
    required = {'stuffstash-cli-release.json', 'stuffstash-selfhost.tar.gz', 'stuffstash-selfhost.tar.gz.sha256'}
    if names != required | {item['name'] for item in cli['assets']} | {item['name'] + '.sha256' for item in cli['assets']}:
        raise ValueError('Unexpected or incomplete publication asset set')
    selfhost = 'stuffstash-selfhost.tar.gz'
    if (root / (selfhost + '.sha256')).read_text() != f'{digest(root / selfhost)}  {selfhost}\n':
        raise ValueError('Invalid self-host checksum')
    return plan


def create_failure(tag, result):
    # Never include stdin/release notes. API errors carry useful validation
    # codes on stdout, while gh's stderr normally supplies the HTTP status.
    details = result.stderr or ''
    try:
        response = json.loads(result.stdout or '{}')
        errors = response.get('errors', [])
        selected = {key: response[key] for key in ('message',) if key in response}
        selected['errors'] = [{key: error[key] for key in ('code', 'field', 'resource') if key in error}
                              for error in errors if isinstance(error, dict)]
        details += ' ' + json.dumps(selected)
    except (ValueError, AttributeError, TypeError):
        pass
    for name in ('GH_TOKEN', 'GITHUB_TOKEN', 'GH_ENTERPRISE_TOKEN', 'GITHUB_ENTERPRISE_TOKEN'):
        token = os.environ.get(name)
        if token:
            details = details.replace(token, '[redacted]')
    details = ''.join(character for character in details if character.isprintable() or character in '\n\t')
    return f'Creating draft release {tag} failed (exit {result.returncode}): {details[:4096]}'


class GitHub:
    def __init__(self, repo, process=subprocess.run):
        if not re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', repo):
            raise ValueError('Invalid repository')
        self.repo = repo
        self.process = process

    def run(self, *args):
        return subprocess.check_output(['gh', *args], text=True)

    def api(self, path):
        return json.loads(self.run('api', f'repos/{self.repo}/{path}'))

    def releases(self):
        # Listing distinguishes absence from an authentication/network failure.
        pages = json.loads(self.run('api', '--paginate', '--slurp', f'repos/{self.repo}/releases?per_page=100'))
        return [release for page in pages for release in page]

    def find(self, tag):
        return next((release for release in self.releases() if release['tag_name'] == tag), None)

    def newer_stable(self, tag):
        version = lambda name: tuple(map(int, TAG.fullmatch(name).groups()))
        return any(not item['draft'] and not item['prerelease'] and TAG.fullmatch(item['tag_name']) and version(item['tag_name']) > version(tag) for item in self.releases())

    def tag_commit(self, tag):
        return self.api('commits/' + tag)['sha']

    def create(self, plan, root):
        body = dict(tag_name=plan['tag'], target_commitish=plan['commit'],
                    name=plan['tag'], body=(root / 'release-notes.md').read_text(), draft=True)
        result = self.process(['gh', 'api', '--method', 'POST',
                                 f'repos/{self.repo}/releases', '--input', '-'],
                                input=json.dumps(body), text=True, capture_output=True, check=False,
                                env={**os.environ, 'GH_DEBUG': ''})
        if result.returncode:
            raise RuntimeError(create_failure(plan['tag'], result))
        return json.loads(result.stdout)

    def refresh(self, release):
        return self.api('releases/' + str(release['id']))

    def download(self, asset, target):
        # Use the authenticated API asset identifier, never an untrusted URL.
        with target.open('wb') as handle:
            subprocess.run(['gh', 'api', '-H', 'Accept: application/octet-stream', f"repos/{self.repo}/releases/assets/{asset['id']}"], stdout=handle, check=True)

    def upload(self, tag, path):
        self.run('release', 'upload', tag, str(path), '--repo', self.repo)

    def publish(self, tag):
        self.run('release', 'edit', tag, '--repo', self.repo, '--draft=false', '--latest=' + ('false' if self.newer_stable(tag) else 'true'))


def publish(root, remote):
    plan = validate_plan(root)
    if remote.tag_commit(plan['tag']) != plan['commit']:
        raise ValueError('Release tag does not point at the original build commit')
    release = remote.find(plan['tag'])
    if release is None:
        release = remote.create(plan, root)
    if release['prerelease']:
        raise ValueError('Refusing to replace a prerelease')
    assets = {asset['name']: asset for asset in release['assets']}
    for expected in plan['assets']:
        name = expected['name']
        if name not in assets:
            remote.upload(plan['tag'], root / name)
        # Fresh read also resolves a previous upload whose response was lost.
        current = remote.refresh(release)
        matches = [item for item in current['assets'] if item['name'] == name]
        if len(matches) != 1:
            raise ValueError('Missing or duplicated remote asset: ' + name)
        with tempfile.TemporaryDirectory() as directory:
            downloaded = Path(directory) / name
            remote.download(matches[0], downloaded)
            if digest(downloaded) != expected['sha256']:
                raise ValueError('Published bytes differ; refusing overwrite: ' + name)
    if release['draft']:
        remote.publish(plan['tag'])
    current = remote.refresh(release)
    if current['draft'] or current['prerelease'] or not current.get('published_at'):
        raise ValueError('Release is not published stable')
    if remote.newer_stable(plan['tag']):
        return None
    cli = json.loads((root / 'stuffstash-cli-release.json').read_text())
    cli.update(repository=remote.repo, publishedAt=current['published_at'])
    return cli


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('directory', type=Path)
    parser.add_argument('--repository', required=True)
    parser.add_argument('--metadata', type=Path, required=True)
    args = parser.parse_args()
    metadata = publish(args.directory, GitHub(args.repository))
    args.metadata.parent.mkdir(parents=True, exist_ok=True)
    args.metadata.write_text(json.dumps(metadata, indent=2) + '\n')


if __name__ == '__main__':
    main()
