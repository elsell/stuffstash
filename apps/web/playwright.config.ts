import { defineConfig, devices } from '@playwright/test';

const port = 5197;

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: {
    timeout: 5_000
  },
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: 'retain-on-failure'
  },
  webServer: {
    command: `vite --host 127.0.0.1 --port ${port}`,
    url: `http://127.0.0.1:${port}`,
    env: {
      VITE_STUFF_STASH_API_BASE_URL: 'http://127.0.0.1:18080',
      VITE_STUFF_STASH_OIDC_ISSUER: 'http://127.0.0.1:5556/dex',
      VITE_STUFF_STASH_OIDC_CLIENT_ID: 'stuff-stash-web-e2e',
      VITE_STUFF_STASH_OIDC_REDIRECT_URI: `http://127.0.0.1:${port}/callback`
    },
    reuseExistingServer: false,
    timeout: 30_000
  },
  projects: [
    {
      name: 'desktop-chromium',
      use: { ...devices['Desktop Chrome'] }
    },
    {
      name: 'mobile-chromium',
      use: { ...devices['Pixel 7'] }
    }
  ]
});
