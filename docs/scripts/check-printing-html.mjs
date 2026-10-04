import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const root = fileURLToPath(new URL('../', import.meta.url));
const manifest = JSON.parse(readFileSync(path.join(root, 'src/assets/printing/generated/manifest.json'), 'utf8'));
const prefix = 'docs/src/content/docs/printing/';
for (const source of Object.keys(manifest.files).filter(name => name.startsWith(prefix) && name.endsWith('.mdx'))) {
  const slug = path.basename(source, '.mdx');
  const markdown = readFileSync(path.join(root, source.slice('docs/'.length)), 'utf8');
  const html = readFileSync(path.join(root, 'dist/printing', slug, 'index.html'), 'utf8');
  const tables = markdown.split('\n').filter(line => /^\|(?:\s*:?-+:?\s*\|)+\s*$/.test(line)).length;
  assert.equal((html.match(/<table\b/g) ?? []).length, tables, `${slug}: generated tables did not render as HTML tables`);
  if (tables) {
    assert.ok(html.includes('<thead>') && /<th[\s>]/.test(html), `${slug}: table column headers are missing`);
  }
  assert.doesNotMatch(html, /<p(?:\s[^>]*)?>\s*\|[^<]*\n\s*\|\s*:?-+/s, `${slug}: raw Markdown table in rendered prose`);
}
console.log('Generated printing pages contain rendered tables, not raw Markdown.');
