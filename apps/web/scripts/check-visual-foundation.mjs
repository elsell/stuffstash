import { readFile, readdir } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const styles = await readFile(new URL('../src/styles.css', import.meta.url), 'utf8');
const packageManifest = JSON.parse(await readFile(new URL('../package.json', import.meta.url), 'utf8'));

async function svelteFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return svelteFiles(path);
    return entry.name.endsWith('.svelte') ? [path] : [];
  }));
  return nested.flat();
}

const requiredTokens = [
  '--space-1: 0.25rem;',
  '--space-2: 0.5rem;',
  '--space-3: 0.75rem;',
  '--space-4: 1rem;',
  '--space-6: 1.5rem;',
  '--space-8: 2rem;',
  '--space-12: 3rem;',
  '--radius-control: 0.5rem;',
  '--radius-surface: 0.75rem;',
  '--radius-overlay: 1rem;',
  '--radius-pill: 999px;',
  '--text-caption-size: 0.75rem;',
  '--text-caption-line-height: 1rem;',
  '--text-metadata-size: 0.8125rem;',
  '--text-metadata-line-height: 1.125rem;',
  '--text-body-size: 0.9375rem;',
  '--text-body-line-height: 1.375rem;',
  '--text-label-size: 0.9375rem;',
  '--text-label-line-height: 1.25rem;',
  '--text-section-size: 1.0625rem;',
  '--text-section-line-height: 1.375rem;',
  '--text-title-size: 1.75rem;',
  '--text-title-line-height: 2.125rem;',
  '--content-max: 72rem;',
  '--content-gutter: 1.5rem;'
];

const missingTokens = requiredTokens.filter((token) => !styles.includes(token));
if (missingTokens.length > 0) {
  throw new Error(`Missing visual-foundation tokens:\n${missingTokens.join('\n')}`);
}

if (styles.includes('@fontsource-variable/inter') || styles.includes('InterVariable')) {
  throw new Error('Task surfaces must use the platform-native font stack, not Inter.');
}

if (packageManifest.dependencies?.['@fontsource-variable/inter'] ||
    packageManifest.devDependencies?.['@fontsource-variable/inter']) {
  throw new Error('The unused Inter package must not be reintroduced; the product uses the platform-native font stack.');
}

if (!styles.includes('font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", sans-serif;')) {
  throw new Error('The platform-native task font stack is missing.');
}

if (!styles.includes('width: min(var(--content-max), calc(100% - (2 * var(--content-gutter))));')) {
  throw new Error('The shared content track is missing.');
}

const mobileRootBlock = styles.match(
  /@media\s*\(max-width:\s*900px\)\s*\{\s*:root\s*\{(?<declarations>[^}]*)\}/
)?.groups?.declarations;
if (!mobileRootBlock?.includes('--text-title-size: 1.5rem;') ||
    !mobileRootBlock.includes('--text-title-line-height: 1.875rem;')) {
  throw new Error('The mobile page-title typography pair must remain 24/30.');
}

const productFiles = await svelteFiles(fileURLToPath(new URL('../src', import.meta.url)));
const productStyles = [styles, ...(await Promise.all(productFiles.map((file) => readFile(file, 'utf8'))))].join('\n');
const supportedBreakpoints = new Set([520, 640, 760, 768, 860, 900, 1024, 1180, 1280]);
const usedBreakpoints = [...productStyles.matchAll(/@media\s*\([^)]*(?:max|min)-width:\s*(\d+)px/g)]
  .map((match) => Number(match[1]));
const unsupportedBreakpoints = [...new Set(usedBreakpoints.filter((value) => !supportedBreakpoints.has(value)))];
if (unsupportedBreakpoints.length > 0) {
  throw new Error(`Undocumented responsive breakpoints: ${unsupportedBreakpoints.join(', ')}px.`);
}
const legacyRawValueCeilings = {
  offScaleSpacing: 73,
  rawFontSize: 120,
  rawRadius: 45,
  rawShadow: 27,
  unsupportedWeight: 25
};
const rawValuePatterns = {
  offScaleSpacing: /(?:gap|padding(?:-[a-z]+)?|margin(?:-[a-z]+)?):[^;\n]*(?:\b6px\b|\b10px\b|\b14px\b)/g,
  rawFontSize: /font-size:\s*[0-9.]+(?:rem|px)/g,
  rawRadius: /border-radius:\s*[0-9.]+(?:rem|px)/g,
  rawShadow: /box-shadow:\s*(?!var\()[^;\n]+/g,
  unsupportedWeight: /font-weight:\s*(?!(?:400|500|600|700)\b)\d{3}/g
};

for (const [category, pattern] of Object.entries(rawValuePatterns)) {
  const count = productStyles.match(pattern)?.length ?? 0;
  const ceiling = legacyRawValueCeilings[category];
  if (count > ceiling) {
    throw new Error(
      `Visual token debt increased for ${category}: ${count} declarations exceeds the ${ceiling} ceiling. ` +
      'Use shared tokens and lower the ceiling when legacy declarations are removed.'
    );
  }
}

console.log('Visual foundation check passed.');
