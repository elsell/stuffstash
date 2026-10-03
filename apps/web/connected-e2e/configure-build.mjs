import { readFile, writeFile } from 'node:fs/promises';
// Preview serves SvelteKit's compiled client; adapter-static copies the same
// artifact into build/. Configure both without changing tracked dev defaults.
for (const location of ['../build/config.json', '../.svelte-kit/output/client/config.json']) {
  const path = new URL(location, import.meta.url);
  const config = JSON.parse(await readFile(path, 'utf8'));
  config.oidcIssuer = 'http://dex:5556/dex';
  await writeFile(path, `${JSON.stringify(config, null, 2)}\n`);
}
