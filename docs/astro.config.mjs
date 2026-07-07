import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import lucode from 'lucode-starlight';

const site = process.env.STUFF_STASH_DOCS_SITE ?? 'https://elsell.github.io';
const base = process.env.STUFF_STASH_DOCS_BASE ?? '/stuffstash/';

export default defineConfig({
  site,
  base,
  integrations: [
    starlight({
      title: 'Stuff Stash',
      logo: {
        src: './src/assets/stuff-stash-glyph.png',
        alt: '',
      },
      favicon: '/brand/stuff-stash-glyph.png',
      customCss: ['./src/styles/brand.css'],
      plugins: [lucode()],
      sidebar: [
        {
          label: 'Evaluate',
          items: [
            { label: 'What It Does', slug: 'product' },
            { label: 'Run Stuff Stash', slug: 'self-hosting' },
            { label: 'Configuration Reference', slug: 'configuration' },
            { label: 'First Inventory', slug: 'first-inventory' },
            { label: 'Concepts', slug: 'concepts' },
            { label: 'Trust And Security', slug: 'security' },
          ],
        },
        {
          label: 'Build',
          items: [
            { label: 'Architecture', slug: 'architecture' },
            { label: 'Development Setup', slug: 'local-development' },
            { label: 'Contributing', slug: 'specs-and-process' },
          ],
        },
      ],
    }),
  ],
});
