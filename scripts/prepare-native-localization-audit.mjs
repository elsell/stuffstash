import { mkdtemp, readFile, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import ts from '../apps/mobile/node_modules/typescript/lib/typescript.js';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const locale = process.argv[2] || 'en';
if (!['en', 'en-XA', 'ar-XB'].includes(locale)) throw new Error('Unsupported native audit locale');
const target = process.argv[3] || path.join(root, 'apps/mobile/native-audit/FixtureAuditTests.swift');
const directory = await mkdtemp(path.join(tmpdir(), 'stuffstash-native-labels-'));
try {
  // Compile the reviewed, dependency-free production formatter; do not duplicate pseudolocalization.
  for (const name of ['mobile', 'web', 'workflow', 'en', 'translator']) {
    const source = await readFile(path.join(root, 'packages/localization/src', `${name}.ts`), 'utf8');
    const result = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } });
    await writeFile(path.join(directory, `${name}.js`), result.outputText);
  }
  const require = createRequire(path.join(directory, 'runtime.cjs'));
  const { en } = require('./en.js');
  const { createTranslator } = require('./translator.js');
  const translator = createTranslator(en, { locale });
  const keys = ['mobile.AddAssetScreen.assetName', 'mobile.AddAssetScreen.addItem', 'mobile.AddAssetScreen.saveItem', 'mobile.AddAssetScreen.closeAdd', 'add.error.save'];
  const entries = keys.map(key => `  ${JSON.stringify(en[key])}: ${JSON.stringify(translator.message(key))}`);
  const source = await readFile(target, 'utf8');
  const marker = /\/\/ AUDIT_LOCALIZATION_LABELS_BEGIN[\s\S]*?\/\/ AUDIT_LOCALIZATION_LABELS_END/;
  if (!marker.test(source)) throw new Error('Missing native audit label marker');
  await writeFile(target, source.replace(marker, `// AUDIT_LOCALIZATION_LABELS_BEGIN\nprivate let auditLocalizationLabels: [String: String] = [\n${entries.join(',\n')}\n]\n// AUDIT_LOCALIZATION_LABELS_END`));
} finally {
  await rm(directory, { recursive: true, force: true });
}
