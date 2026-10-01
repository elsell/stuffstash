import { writeFile } from 'node:fs/promises';
import { expect, test } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

type Measurement = {
  platform: string; operation: string; surface: string; variant: string;
  outcome: string; durationMs: number;
};

for (const decode of ['success', 'failure'] as const) {
  test(`visible image telemetry reports real browser decode ${decode}`, async ({ page }, testInfo) => {
    resetWorkspaceApiState(page);
    await installAuthenticatedWorkspace(page);
    const samples: Measurement[] = [];
    const authorizations: (string | undefined)[] = [];
    await page.route('**/config.json', route => route.fulfill({ json: {
      apiBaseUrl: 'http://127.0.0.1:18080', oidcIssuer: 'http://127.0.0.1:5556/dex',
      oidcClientId: 'stuff-stash-web-e2e', oidcRedirectUri: 'http://127.0.0.1:5197/callback',
      performanceTelemetryEnabled: true
    } }));
    await page.route('**/client-telemetry', async route => {
      authorizations.push(route.request().headers().authorization);
      const batch = route.request().postDataJSON().measurements as Measurement[];
      samples.push(...batch);
      await route.fulfill({ json: { data: { accepted: batch.length } } });
    });
    await page.route(/\/assets\/asset-tomato\/attachments(?:\?.*)?$/, route => route.fulfill({ json: { data: [{
      id: 'attachment-photo', tenantId: 'tenant-home', inventoryId: 'inventory-household',
      assetId: 'asset-tomato', fileName: 'photo.png', contentType: 'image/png', sizeBytes: 1024, lifecycleState: 'active'
    }], meta: { pagination: { limit: 50, nextCursor: null, hasMore: false } } } }));
    if (decode === 'failure') {
      await page.route('**/thumbnail?*', route => route.fulfill({ contentType: 'image/png', body: 'invalid-image' }));
    }
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'Home', exact: true })).toBeVisible();
    await expect.poll(() => samples.filter(s => s.operation === 'image' && s.surface === 'home' && s.outcome === decode).length).toBeGreaterThan(0);
    await page.getByRole('link', { name: /Tomato fertilizer/ }).first().click();
    await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
    await expect.poll(() => samples.filter(s => s.operation === 'image' && s.surface === 'detail' && s.outcome === decode).length).toBeGreaterThan(0);
    expect(authorizations.length).toBeGreaterThan(0);
    expect(authorizations.every(value => value === 'Bearer e2e-token')).toBe(true);
    for (const sample of samples) {
      expect(Object.keys(sample).sort()).toEqual(['durationMs', 'operation', 'outcome', 'platform', 'surface', 'variant']);
      expect(sample.platform).toBe('web');
      expect(['request', 'image']).toContain(sample.operation);
      expect(['application', 'home', 'list', 'detail', 'gallery', 'fullscreen', 'upload']).toContain(sample.surface);
      expect(['none', 'small', 'medium', 'large', 'original']).toContain(sample.variant);
      expect(['success', 'failure', 'cancelled']).toContain(sample.outcome);
      expect(Number.isFinite(sample.durationMs)).toBe(true);
      expect(sample.durationMs).toBeGreaterThanOrEqual(0);
      expect(sample.durationMs).toBeLessThanOrEqual(60000);
    }
    const evidencePath = testInfo.outputPath('visible-image-samples.json');
    await writeFile(evidencePath, JSON.stringify({
      evidence: 'Controlled browser decode and telemetry delivery; not production latency or real OIDC.',
      decode, browser: testInfo.project.name, samples: samples.filter(s => s.operation === 'image')
    }, null, 2));
    await testInfo.attach('visible-image-samples.json', { path: evidencePath, contentType: 'application/json' });
  });
}
