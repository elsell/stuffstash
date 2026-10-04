#!/usr/bin/env python3
import importlib.util
from pathlib import Path
import unittest
import os
import re
import subprocess
import tempfile
import textwrap

spec = importlib.util.spec_from_file_location('pages_config', Path(__file__).with_name('configure-pages.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)

class PagesConfigTests(unittest.TestCase):
    def test_custom_domain_production_and_preview(self):
        pages = {'html_url': 'https://stuffstash.org/', 'cname': 'stuffstash.org'}
        self.assertEqual(module.configuration(pages), {'site': 'https://stuffstash.org', 'base': '/', 'cname': 'stuffstash.org', 'url': 'https://stuffstash.org/'})
        self.assertEqual(module.configuration(pages, 405)['base'], '/pr-405/')
        self.assertEqual(module.configuration(pages, 405)['url'], 'https://stuffstash.org/pr-405/')

    def test_project_path_production_and_preview(self):
        pages = {'html_url': 'https://elsell.github.io/stuffstash/', 'cname': None}
        self.assertEqual(module.configuration(pages)['base'], '/stuffstash/')
        self.assertEqual(module.configuration(pages, 405), {'site': 'https://elsell.github.io', 'base': '/stuffstash/pr-405/', 'cname': '', 'url': 'https://elsell.github.io/stuffstash/pr-405/'})

    def test_invalid_public_config_cannot_inject_workflow_environment(self):
        for url in ['http://stuffstash.org/', 'https://user@stuffstash.org/', 'https://stuffstash.org/?x=1', 'https://stuffstash.org/\nOTHER=value']:
            with self.subTest(url=url), self.assertRaises(ValueError):
                module.configuration({'html_url': url})
        with self.assertRaises(ValueError):
            module.configuration({'html_url':'https://stuffstash.org/', 'cname':'other.example'})

class PagesPublicationTests(unittest.TestCase):
    def test_production_content_preserves_configured_domain_and_previews(self):
        workflow = (Path(__file__).parents[1] / '.github/workflows/docs-pages.yml').read_text()
        step = re.search(r'      - name: Publish production docs\n(.*?)(?=\n      - name:)', workflow, re.S).group(1)
        command = textwrap.dedent(step.split('        run: |\n', 1)[1])
        for cname in ['stuffstash.org', '']:
            with self.subTest(cname=cname), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                (root / 'pages/pr-7').mkdir(parents=True)
                (root / 'pages/pr-7/index.html').write_text('Keep preview')
                (root / 'pages/CNAME').write_text('previous.example')
                (root / 'pages/old.html').write_text('Stale production')
                (root / 'built-docs').mkdir()
                (root / 'built-docs/index.html').write_text('Current production')
                subprocess.run(['bash', '-e', '-c', command], cwd=root,
                               env={**os.environ, 'PAGES_CNAME': cname}, check=True)
                self.assertEqual((root / 'pages/index.html').read_text(), 'Current production')
                self.assertEqual((root / 'pages/pr-7/index.html').read_text(), 'Keep preview')
                self.assertFalse((root / 'pages/old.html').exists())
                self.assertTrue((root / 'pages/.nojekyll').exists())
                if cname:
                    self.assertEqual((root / 'pages/CNAME').read_text(), cname + '\n')
                else:
                    self.assertFalse((root / 'pages/CNAME').exists())


if __name__ == '__main__': unittest.main()
