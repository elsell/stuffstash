import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import lucode from 'lucode-starlight';

const site = process.env.STUFF_STASH_DOCS_SITE ?? 'https://elsell.github.io';
const base = process.env.STUFF_STASH_DOCS_BASE ?? '/stuffstash/';

export default defineConfig({
  site,
  base,
  // MDX must receive an explicit value; Astro’s Markdown processor default is not inherited.
  markdown: { gfm: true },
  integrations: [
    starlight({
      title: 'Stuff Stash',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/elsell/stuffstash' },
      ],
      logo: {
        src: './src/assets/stuff-stash-glyph.png',
        alt: '',
      },
      editLink: {
        baseUrl: 'https://github.com/elsell/stuffstash/edit/main/docs/',
      },
      favicon: '/brand/stuff-stash-glyph.png',
      customCss: ['./src/styles/brand.css'],
      plugins: [lucode()],
      sidebar: [
        { label: 'Printing', items: [
          { label: 'Printing labels', slug: 'printing/setup' },
          { label: 'Supported printers', slug: 'printing/supported-printers' },
          { label: 'Label sizes', slug: 'printing/label-sizes' },
          { label: 'Templates', slug: 'printing/templates' },
        ] },
        {
          label: 'Evaluate',
          items: [
            { label: 'What It Does', slug: 'product' },
            { label: 'Run Stuff Stash', slug: 'self-hosting' },
            { label: 'Self-Host Operations', slug: 'self-host-operations' },
            { label: 'Dex Users And Clients', slug: 'dex-users' },
            { label: 'Configuration Reference', slug: 'configuration' },
            { label: 'First Inventory', slug: 'first-inventory' },
            { label: 'Expiration Dates', slug: 'expiration' },
            { label: 'Export An Inventory', slug: 'export-inventory' },
            { label: 'Ask About Your Inventory', slug: 'conversation' },
            { label: 'Concepts', slug: 'concepts' },
            { label: 'Trust And Security', slug: 'security' },
          ],
        },
        {
          label: 'Build',
          items: [
            { label: 'Architecture', slug: 'architecture' },
            { label: 'Development Setup', slug: 'local-development' },
            {
              label: 'Command line',
              items: [
                { label: 'Download and verify', slug: 'cli-downloads' },
                { label: 'Sign in and use inventories', slug: 'cli' },
                { label: 'Printers and labels', slug: 'cli-printing' },
                { label: 'Administration and backups', slug: 'cli-administration' },
              ],
            },
            { label: 'Connect An Inventory Agent', slug: 'mcp' },
            { label: 'Compatible Language Providers', slug: 'compatible-providers' },
            { label: 'Release To TestFlight', slug: 'testflight' },
            { label: 'Store Release Notes', slug: 'store-release-notes' },
            { label: 'Contributing', slug: 'specs-and-process' },
          ],
        },
      ],
    }),
  ],
});
