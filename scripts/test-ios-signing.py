"""Signing-input contract tests; synthetic profiles, no Apple credentials."""
import datetime
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('signing', Path(__file__).with_name('ios-signing.py'))
signing = importlib.util.module_from_spec(spec)
spec.loader.exec_module(signing)

class SigningTests(unittest.TestCase):
    def profile(self):
        return {'UUID': '12345678-1234-1234-1234-123456789abc',
                'TeamIdentifier': ['TEAM123'],
                'ApplicationIdentifierPrefix': ['PREFIX'],
                'ExpirationDate': datetime.datetime(2030, 1, 1),
                'Entitlements': {'application-identifier': 'PREFIX.org.example.app',
                                 'com.apple.developer.team-identifier': 'TEAM123',
                                 'get-task-allow': False},
                'DeveloperCertificates': [b'certificate']}

    def validate(self, profile):
        return signing.validate_profile(profile, 'TEAM123', 'org.example.app',
                                        datetime.datetime(2026, 1, 1))

    def test_distribution_profile(self):
        self.assertEqual(self.validate(self.profile()), '12345678-1234-1234-1234-123456789abc')

    def test_rejects_unsafe_profiles(self):
        for change in ({'ExpirationDate': datetime.datetime(2025, 1, 1)},
                       {'TeamIdentifier': ['OTHER']}, {'ProvisionedDevices': ['device']},
                       {'ProvisionsAllDevices': True}, {'UUID': '../escape'},
                       {'DeveloperCertificates': []}):
            with self.subTest(change=change), self.assertRaises(ValueError):
                self.validate(self.profile() | change)
        for key, value in [('application-identifier', 'PREFIX.other'),
                           ('get-task-allow', True),
                           ('com.apple.developer.team-identifier', 'OTHER')]:
            profile = self.profile()
            profile['Entitlements'][key] = value
            with self.subTest(key=key), self.assertRaises(ValueError):
                self.validate(profile)

    def test_identity_must_match_profile(self):
        import hashlib
        fingerprint = hashlib.sha1(b'certificate').hexdigest().upper()
        identities = f'1) {fingerprint} "Apple Distribution: Example"'
        self.assertEqual(signing.select_identity(self.profile(), identities), fingerprint)
        with self.assertRaises(ValueError):
            signing.select_identity(self.profile(), '0 valid identities found')
        with self.assertRaises(ValueError):
            signing.select_identity(self.profile(), f'1) {fingerprint} "Apple Development: Example"')

if __name__ == '__main__':
    unittest.main()
