import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://stuffstash.online',
  integrations: [
    starlight({
      title: 'Stuff Stash',
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
