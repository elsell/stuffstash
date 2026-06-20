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
        alt: 'Stuff Stash',
      },
      favicon: '/brand/stuff-stash-glyph.png',
      customCss: ['./src/styles/brand.css'],
      sidebar: [
        {
          label: 'Start Here',
          items: [
            { label: 'Overview', slug: 'overview' },
            { label: 'Local Development', slug: 'local-development' },
            { label: 'Architecture', slug: 'architecture' },
            { label: 'Specs And Process', slug: 'specs-and-process' },
          ],
        },
      ],
    }),
  ],
});
