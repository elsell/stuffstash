import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './connected-e2e',
  outputDir: './test-results/connected',
  timeout: 90_000,
  expect: { timeout: 20_000 },
  workers: 1,
  use: { baseURL: 'http://localhost:5173', trace: 'off', video: 'off', screenshot: 'off' },
  projects: [{ name: 'connected-chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: 'vite --host 0.0.0.0 --port 5173 --strictPort',
    url: 'http://localhost:5173', reuseExistingServer: false,
    env: {
      VITE_STUFF_STASH_API_BASE_URL: 'http://localhost:8080',
      VITE_STUFF_STASH_OIDC_ISSUER: 'http://dex:5556/dex',
      VITE_STUFF_STASH_OIDC_CLIENT_ID: 'stuff-stash-web-local',
      VITE_STUFF_STASH_OIDC_REDIRECT_URI: 'http://localhost:5173/callback'
    }
  }
});
