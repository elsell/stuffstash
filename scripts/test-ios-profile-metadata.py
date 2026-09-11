import datetime
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('metadata', Path(__file__).with_name('ios-profile-metadata.py'))
metadata = importlib.util.module_from_spec(spec)
spec.loader.exec_module(metadata)
fixture_spec = importlib.util.spec_from_file_location('fixtures', Path(__file__).with_name('test-ios-signing.py'))
fixtures = importlib.util.module_from_spec(fixture_spec)
fixture_spec.loader.exec_module(fixtures)

class MetadataTests(unittest.TestCase):
    def validate(self, profile, original=None):
        return metadata.profile_metadata(profile, 'TEAM123', 'org.example.app',
            datetime.datetime(2026, 1, 1), original=original)

    def test_valid_replacement_keeps_certificate(self):
        profile = fixtures.SigningTests().profile()
        original = self.validate(profile)
        self.assertEqual(self.validate(profile, original), original)

    def test_replacement_rejects_wrong_scope_certificate_and_missing_push(self):
        profile = fixtures.SigningTests().profile()
        original = self.validate(profile)
        for change in [{'TeamIdentifier': ['OTHER']}, {'DeveloperCertificates': [b'other']},
                       {'ExpirationDate': datetime.datetime(2025, 1, 1)}]:
            with self.assertRaises(ValueError):
                self.validate(profile | change, original)
        del profile['Entitlements']['aps-environment']
        self.validate(profile)
        with self.assertRaises(ValueError):
            self.validate(profile, original)

    def test_replacement_rejects_associated_domains(self):
        profile = fixtures.SigningTests().profile()
        original = self.validate(profile)
        profile['Entitlements']['com.apple.developer.associated-domains'] = ['applinks:example.com']
        with self.assertRaises(ValueError):
            self.validate(profile, original)

if __name__ == '__main__':
    unittest.main()
