import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './connected-e2e',
  outputDir: './test-results/connected',
  timeout: 90_000,
  expect: { timeout: 20_000 },
  workers: 1,
  use: { baseURL: 'http://localhost:5173', actionTimeout: 20_000, trace: 'off', video: 'off', screenshot: 'off' },
  projects: [{ name: 'connected-chromium', use: { ...devices['Desktop Chrome'], launchOptions: { args: ['--host-resolver-rules=MAP dex 127.0.0.1'] } } }],
  webServer: {
    command: 'pnpm build && node connected-e2e/configure-build.mjs && vite preview --host 0.0.0.0 --port 5173 --strictPort',
    timeout: 180_000,
    url: 'http://localhost:5173', reuseExistingServer: false
  }
});
