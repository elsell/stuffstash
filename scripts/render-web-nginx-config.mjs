import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';

const [indexPath, templatePath, outputPath] = process.argv.slice(2);

if (!indexPath || !templatePath || !outputPath) {
  throw new Error('usage: render-web-nginx-config.mjs <index.html> <nginx-template> <output>');
}

const index = readFileSync(indexPath, 'utf8');
const template = readFileSync(templatePath, 'utf8');
const match = index.match(/<script>([\s\S]*?)<\/script>/);

if (!match) {
  throw new Error(`unable to find SvelteKit bootstrap script in ${indexPath}`);
}

const hash = createHash('sha256').update(match[1], 'utf8').digest('base64');
const rendered = template.replace('__STUFFSTASH_BOOTSTRAP_SCRIPT_HASH__', `sha256-${hash}`);

if (rendered === template) {
  throw new Error(`unable to find CSP hash placeholder in ${templatePath}`);
}

writeFileSync(outputPath, rendered);
