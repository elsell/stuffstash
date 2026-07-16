import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { loadEnv, type Plugin } from 'vite';
import { defineConfig } from 'vitest/config';

import { resolveDevRuntimeConfig } from './src/lib/devRuntimeConfig';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_');
  return {
    plugins: [devRuntimeConfigPlugin(env), tailwindcss(), sveltekit()],
    server: {
      proxy: env.VITE_STUFF_STASH_PROXY_DEX === 'true'
        ? {
            '/dex': {
              target: env.VITE_STUFF_STASH_DEX_PROXY_TARGET || 'http://127.0.0.1:5556',
              changeOrigin: true
            }
          }
        : undefined
    },
    resolve: {
      conditions: ['browser']
    },
    test: {
      environment: 'jsdom',
      include: ['src/**/*.test.ts'],
      setupFiles: ['./src/test/setup.ts'],
      maxWorkers: 4
    }
  };
});

function devRuntimeConfigPlugin(env: Record<string, string>): Plugin {
  return {
    name: 'stuffstash-dev-runtime-config',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/config.json', (request, response) => {
        const configPath = resolve(server.config.root, 'static/config.json');
        const config = resolveDevRuntimeConfig(
          JSON.parse(readFileSync(configPath, 'utf8')),
          env,
          request.headers.host
        );
        response.setHeader('Cache-Control', 'no-store');
        response.setHeader('Content-Type', 'application/json');
        response.end(`${JSON.stringify(config, null, 2)}\n`);
      });
    }
  };
}
