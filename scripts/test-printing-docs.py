#!/usr/bin/env python3
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

spec=importlib.util.spec_from_file_location('generator',Path(__file__).with_name('generate-printing-docs.py'))
generator=importlib.util.module_from_spec(spec);spec.loader.exec_module(generator)


class GeneratedOwnershipTests(unittest.TestCase):
    def files(self, images):
        files = {str(generator.ASSETS/name): content for name,content in images.items()}
        files[str(generator.MANIFEST)] = json.dumps({'files':{path:'digest' for path in files}}).encode()
        return files

    def test_binary_drift_missing_extra_and_removed_outputs_preserve_curated_docs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp); guide=root/generator.PAGES/'setup.md'; guide.parent.mkdir(parents=True);guide.write_text('Human guide')
            first=self.files({'old.png':b'first raster'})
            generator.apply(first,root,False)
            self.assertEqual(generator.apply(first,root,True),[])
            (root/generator.ASSETS/'old.png').write_bytes(b'changed pixel')
            self.assertIn(str(generator.ASSETS/'old.png'),generator.apply(first,root,True))
            second=self.files({'new.png':b'new template raster'})
            generator.apply(second,root,False)
            self.assertFalse((root/generator.ASSETS/'old.png').exists())
            self.assertEqual(guide.read_text(),'Human guide')
            (root/generator.ASSETS/'unknown.png').write_bytes(b'unknown owner')
            self.assertIn(str(generator.ASSETS/'unknown.png'),generator.apply(second,root,True))
            generator.apply(second,root,False)
            self.assertTrue((root/generator.ASSETS/'unknown.png').exists())
            (root/generator.ASSETS/'new.png').unlink()
            self.assertIn(str(generator.ASSETS/'new.png'),generator.apply(second,root,True))

    def test_poisoned_manifest_cannot_delete_curated_content(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);guide=root/generator.PAGES/'setup.md';guide.parent.mkdir(parents=True);guide.write_text('Keep')
            manifest=root/generator.MANIFEST;manifest.parent.mkdir(parents=True)
            manifest.write_text(json.dumps({'files':{str(generator.PAGES/'setup.md'):'anything'}}))
            with self.assertRaisesRegex(ValueError,'unowned'):
                generator.apply(self.files({}),root,False)
            self.assertEqual(guide.read_text(),'Keep')


if __name__=='__main__':unittest.main()
