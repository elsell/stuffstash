"""Decode and validate maintenance profiles without exposing signing contents."""
import base64
import datetime
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import plistlib
import subprocess
import sys

spec = importlib.util.spec_from_file_location('signing', Path(__file__).with_name('ios-signing.py'))
signing = importlib.util.module_from_spec(spec)
spec.loader.exec_module(signing)


def profile_metadata(profile, team, bundle, now, *, original=None):
    signing.validate_profile(profile, team, bundle, now, require_push=original is not None)
    certificates = [hashlib.sha256(value).hexdigest() for value in profile['DeveloperCertificates']]
    if original is not None:
        if original['teamId'] != team or original['bundleId'] != bundle:
            raise ValueError('Original profile scope mismatch')
        if len(certificates) != 1 or certificates[0] not in original['certificateSha256']:
            raise ValueError('Replacement profile changed distribution certificate')
        if profile['Entitlements'].get('com.apple.developer.associated-domains'):
            raise ValueError('General distribution profile unexpectedly enables associated domains')
    return {'teamId': team, 'bundleId': bundle, 'certificateSha256': certificates}


def main():
    mode, directory = sys.argv[1:]
    directory = Path(directory)
    options = plistlib.loads(Path('apps/mobile/ios/ExportOptions-TestFlight.plist').read_bytes())
    now = datetime.datetime.now(datetime.timezone.utc).replace(tzinfo=None)
    if mode == 'prepare':
        content = base64.b64decode(os.environ['BUILD_PROVISION_PROFILE_BASE64'], validate=True)
        path = directory / 'original.mobileprovision'
        path.write_bytes(content)
        path.chmod(0o600)
        original = None
    elif mode == 'validate':
        path = directory / 'replacement.mobileprovision'
        original = json.loads((directory / 'original.json').read_text())
    else:
        raise ValueError('Invalid maintenance mode')
    result = subprocess.run(['security', 'cms', '-D', '-i', str(path)], capture_output=True)
    if result.returncode:
        raise ValueError('Could not decode provisioning profile')
    metadata = profile_metadata(plistlib.loads(result.stdout), options['teamID'],
                                'org.stuffstash.mobile', now, original=original)
    if mode == 'prepare':
        (directory / 'original.json').write_text(json.dumps(metadata))
    else:
        (directory / 'validated').touch()


if __name__ == '__main__':
    try:
        main()
    except Exception:
        raise SystemExit('Provisioning profile maintenance validation failed') from None
