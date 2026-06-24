import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

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
      sidebar: [
        {
          label: 'Evaluate',
          items: [
            { label: 'Product', slug: 'product' },
            { label: 'Self-Hosting', slug: 'self-hosting' },
            { label: 'Security', slug: 'security' },
            { label: 'Architecture', slug: 'architecture' },
          ],
        },
        {
          label: 'Build',
          items: [
            { label: 'Development Setup', slug: 'local-development' },
            { label: 'Contributing', slug: 'specs-and-process' },
          ],
        },
      ],
    }),
  ],
});
