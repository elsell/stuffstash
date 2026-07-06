import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { loadEnv, type Plugin } from 'vite';
import { defineConfig } from 'vitest/config';

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
      include: ['src/**/*.test.ts']
    }
  };
});

function devRuntimeConfigPlugin(env: Record<string, string>): Plugin {
  return {
    name: 'stuffstash-dev-runtime-config',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/config.json', (_request, response) => {
        const configPath = resolve(server.config.root, 'static/config.json');
        const config = applyRuntimeConfigOverrides(JSON.parse(readFileSync(configPath, 'utf8')), env);
        response.setHeader('Cache-Control', 'no-store');
        response.setHeader('Content-Type', 'application/json');
        response.end(`${JSON.stringify(config, null, 2)}\n`);
      });
    }
  };
}

function applyRuntimeConfigOverrides(config: Record<string, unknown>, env: Record<string, string>): Record<string, unknown> {
  const webOrigin = normalizedOrigin(env.VITE_STUFF_STASH_WEB_ORIGIN);
  return {
    ...config,
    apiBaseUrl: trimTrailingSlash(env.VITE_STUFF_STASH_API_BASE_URL || (webOrigin ? originWithPort(webOrigin, '8080') : String(config.apiBaseUrl))),
    oidcIssuer: trimTrailingSlash(env.VITE_STUFF_STASH_OIDC_ISSUER || (webOrigin ? `${originWithPort(webOrigin, '5556')}/dex` : String(config.oidcIssuer))),
    oidcClientId: env.VITE_STUFF_STASH_OIDC_CLIENT_ID || config.oidcClientId,
    oidcRedirectUri: env.VITE_STUFF_STASH_OIDC_REDIRECT_URI || (webOrigin ? `${webOrigin}/callback` : config.oidcRedirectUri)
  };
}

function normalizedOrigin(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? new URL(trimmed).origin : undefined;
}

function originWithPort(origin: string, port: string): string {
  const url = new URL(origin);
  url.port = port;
  return url.origin;
}

function trimTrailingSlash(value: string): string {
  return value.trim().replace(/\/+$/, '');
}
