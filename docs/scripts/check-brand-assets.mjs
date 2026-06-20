import { readFile } from 'node:fs/promises';

const publicLogoPath = new URL('../public/brand/stuff-stash-glyph.png', import.meta.url);
const sourceLogoPath = new URL('../src/assets/stuff-stash-glyph.png', import.meta.url);

const [publicLogo, sourceLogo] = await Promise.all([
  readFile(publicLogoPath),
  readFile(sourceLogoPath),
]);

if (!publicLogo.equals(sourceLogo)) {
  console.error('Brand logo assets drifted. Keep docs/public/brand/stuff-stash-glyph.png and docs/src/assets/stuff-stash-glyph.png byte-identical.');
  process.exit(1);
}
