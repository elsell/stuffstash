import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const root = new URL('..', import.meta.url).pathname;

const requiredDependencies = {
  '@internationalized/date': '3.12.2',
  '@lucide/svelte': '1.17.0',
  'bits-ui': '2.18.1',
  'class-variance-authority': '0.7.1',
  clsx: '2.1.1',
  'tailwind-merge': '3.6.0',
  'tailwind-variants': '3.2.2'
};

const requiredDevDependencies = {
  '@tailwindcss/vite': '4.3.0',
  'shadcn-svelte': '1.3.0',
  tailwindcss: '4.3.0',
  'tw-animate-css': '1.4.0'
};

const requiredComponentFiles = [
  'components.json',
  'src/lib/utils.ts',
  'src/lib/components/ui/alert/index.ts',
  'src/lib/components/ui/badge/index.ts',
  'src/lib/components/ui/button/index.ts',
  'src/lib/components/ui/card/index.ts',
  'src/lib/components/ui/input/index.ts',
  'src/lib/components/ui/label/index.ts',
  'src/lib/components/ui/select/index.ts',
  'src/lib/components/ui/separator/index.ts',
  'src/lib/components/ui/tabs/index.ts',
  'src/lib/components/ui/textarea/index.ts'
];

const primitivePatterns = [
  /<button\b/,
  /<input\b/,
  /<select\b/,
  /<textarea\b/,
  /<label\b/,
  /class="[^"]*\bsecondary\b/,
  /class="[^"]*\bdanger-button\b/,
  /class="[^"]*\bsegmented\b/
];

const packageJson = readJson('package.json');
assertExactVersions(packageJson.dependencies ?? {}, requiredDependencies, 'dependencies');
assertExactVersions(packageJson.devDependencies ?? {}, requiredDevDependencies, 'devDependencies');

const components = readJson('components.json');
assertEqual(components.registry, 'https://shadcn-svelte.com/registry', 'components.json registry');
assertEqual(components.style, 'vega', 'components.json style');
assertEqual(components.iconLibrary, 'lucide', 'components.json icon library');
assertEqual(components.aliases?.ui, '$lib/components/ui', 'components.json UI alias');

for (const file of requiredComponentFiles) {
  if (!existsSync(join(root, file))) {
    fail(`missing shadcn foundation file: ${file}`);
  }
}

for (const file of svelteFiles(join(root, 'src'))) {
  const normalized = relative(root, file);
  if (normalized.startsWith('src/lib/components/ui/')) {
    continue;
  }
  const contents = readFileSync(file, 'utf8');
  for (const pattern of primitivePatterns) {
    if (pattern.test(contents)) {
      fail(`raw generic primitive detected in ${normalized}: ${pattern}`);
    }
  }
}

function readJson(path) {
  return JSON.parse(readFileSync(join(root, path), 'utf8'));
}

function assertExactVersions(actual, expected, label) {
  for (const [name, version] of Object.entries(expected)) {
    assertEqual(actual[name], version, `${label}.${name}`);
  }
}

function assertEqual(actual, expected, label) {
  if (actual !== expected) {
    fail(`${label} expected ${expected}, got ${actual ?? '<missing>'}`);
  }
}

function svelteFiles(directory) {
  const files = [];
  for (const entry of readdirSync(directory)) {
    const path = join(directory, entry);
    const stats = statSync(path);
    if (stats.isDirectory()) {
      files.push(...svelteFiles(path));
    } else if (path.endsWith('.svelte')) {
      files.push(path);
    }
  }
  return files;
}

function fail(message) {
  console.error(message);
  process.exit(1);
}
