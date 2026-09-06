#!/usr/bin/env python3
"""Install reusable CI signing inputs without requesting Apple certificates."""
import base64
import datetime
import hashlib
import os
from pathlib import Path
import plistlib
import re
import secrets
import subprocess


def validate_profile(profile, team, bundle, now):
    entitlements = profile.get('Entitlements', {})
    prefixes = profile.get('ApplicationIdentifierPrefix', [])
    if (profile.get('TeamIdentifier') != [team]
            or entitlements.get('com.apple.developer.team-identifier') != team
            or entitlements.get('application-identifier') not in [f'{p}.{bundle}' for p in prefixes]
            or entitlements.get('get-task-allow') is not False
            or 'ProvisionedDevices' in profile or profile.get('ProvisionsAllDevices')
            or profile.get('ExpirationDate', now) <= now
            or not profile.get('DeveloperCertificates')
            or not re.fullmatch(r'[0-9a-fA-F-]{36}', profile.get('UUID', ''))):
        raise ValueError('Expected an unexpired App Store Connect profile for the release team and bundle')
    return profile['UUID']


def select_identity(profile, identities):
    for certificate in profile['DeveloperCertificates']:
        fingerprint = hashlib.sha1(certificate).hexdigest().upper()
        if re.search(rf'\b{fingerprint} "Apple Distribution:', identities):
            return fingerprint
    raise ValueError('Profile has no matching valid Apple Distribution identity with a private key')


def command(*args):
    # Never include subprocess arguments (which can contain secrets) in failures.
    result = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if result.returncode:
        raise ValueError(f'Signing operation failed: {args[0]} {args[1]}')
    return result.stdout


def install():
    temporary = Path(os.environ['RUNNER_TEMP'])
    directory = temporary / 'stuffstash-signing'
    directory.mkdir(mode=0o700)
    for variable, name in [('BUILD_CERTIFICATE_BASE64', 'identity.p12'),
                           ('BUILD_PROVISION_PROFILE_BASE64', 'profile.mobileprovision')]:
        content = base64.b64decode(os.environ[variable], validate=True)
        path = directory / name
        path.touch(mode=0o600, exist_ok=False)
        path.write_bytes(content)
    profile = plistlib.loads(command('security', 'cms', '-D', '-i', str(directory / 'profile.mobileprovision')))
    options_path = Path('apps/mobile/ios/ExportOptions-TestFlight.plist')
    options = plistlib.loads(options_path.read_bytes())
    team = options['teamID']
    bundle = 'org.stuffstash.mobile'
    uuid = validate_profile(profile, team, bundle, datetime.datetime.now(datetime.timezone.utc).replace(tzinfo=None))
    keychain = str(directory / 'signing.keychain-db')
    password = secrets.token_urlsafe(32)
    command('security', 'create-keychain', '-p', password, keychain)
    command('security', 'set-keychain-settings', '-lut', '21600', keychain)
    command('security', 'unlock-keychain', '-p', password, keychain)
    command('security', 'import', str(directory / 'identity.p12'), '-P', os.environ['P12_PASSWORD'],
            '-k', keychain, '-T', '/usr/bin/codesign', '-T', '/usr/bin/security')
    command('security', 'set-key-partition-list', '-S', 'apple-tool:,apple:,codesign:', '-k', password, keychain)
    # Preserve the runner's existing search list for system identities.
    existing = command('security', 'list-keychains', '-d', 'user').decode()
    import shlex
    command('security', 'list-keychains', '-d', 'user', '-s', keychain, *shlex.split(existing))
    identity = select_identity(profile, command('security', 'find-identity', '-v', '-p', 'codesigning', keychain).decode())
    profiles = Path.home() / 'Library/Developer/Xcode/UserData/Provisioning Profiles'
    profiles.mkdir(parents=True, exist_ok=True)
    installed = profiles / f'{uuid}.mobileprovision'
    installed.touch(mode=0o600, exist_ok=False)
    installed.write_bytes((directory / 'profile.mobileprovision').read_bytes())
    (directory / 'installed-profile-path').write_text(str(installed))
    options.update(signingStyle='manual', signingCertificate=identity, provisioningProfiles={bundle: uuid})
    (directory / 'ExportOptions.plist').write_bytes(plistlib.dumps(options))
    with open(os.environ['GITHUB_ENV'], 'a') as environment:
        environment.write(f'STUFF_STASH_SIGNING_IDENTITY={identity}\nSTUFF_STASH_SIGNING_PROFILE={uuid}\n')
    (directory / 'identity.p12').unlink()


if __name__ == '__main__':
    try:
        install()
    except (ValueError, KeyError, OSError) as error:
        raise SystemExit(str(error)) from None
