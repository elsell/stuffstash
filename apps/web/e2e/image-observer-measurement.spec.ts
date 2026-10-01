import { writeFile } from 'node:fs/promises';
import { expect, test } from '@playwright/test';

test('measure bounded cached-image observer overhead', async ({ page }, testInfo) => {
  await page.goto('/');
  const evidence = await page.evaluate(async () => {
    // Vite serves the actual production modules; no copy of the measured implementation.
    const actionPath = '/src/lib/presentation/observability/visibleImage.ts';
    const sessionPath = '/src/lib/adapters/observability/webPerformanceSession.ts';
    const { visibleImage } = await import(/* @vite-ignore */ actionPath);
    const { createWebPerformanceSession } = await import(/* @vite-ignore */ sessionPath);
    const canvas = document.createElement('canvas');
    canvas.width = canvas.height = 512;
    const ctx = canvas.getContext('2d')!;
    for (let y = 0; y < 512; y += 8) {
      for (let x = 0; x < 512; x += 8) {
        ctx.fillStyle = `rgb(${x % 256},${y % 256},${(x + y) % 256})`;
        ctx.fillRect(x, y, 8, 8);
      }
    }
    const img = new Image();
    img.src = canvas.toDataURL('image/png');
    document.body.appendChild(img);
    await img.decode();
    const decoded = img.complete && img.naturalWidth === 512;
    let reported = 0;
    let failed = 0;
    let activeDeliveries = 0;
    const scheduled = new Set<() => void>();
    const sessions = [false, true].map(enabled => createWebPerformanceSession({
      enabled, baseUrl: location.origin,
      tokenProvider: async () => 'controlled-measurement',
      scheduler: { schedule: (callback: () => void, delayMs: number) => {
        if (delayMs === 10000) {
          activeDeliveries++;
          const timer = setTimeout(callback, delayMs);
          return () => { clearTimeout(timer); activeDeliveries--; };
        }
        scheduled.add(callback);
        return () => scheduled.delete(callback);
      } },
      fetch: async (input: Request) => {
        const values = (await input.json()).measurements;
        reported += values.length;
        failed += values.filter((value: { outcome: string }) => value.outcome !== 'success').length;
        return new Response(JSON.stringify({ data: { accepted: values.length } }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }));
    const perBatch = 20;
    const samples: { pair: number; enabled: boolean; durationMs: number; perLifecycleMs: number }[] = [];
    function batch(enabled: boolean) {
      const start = performance.now();
      for (let n = 0; n < perBatch; n++) {
        const action = visibleImage(img, {
          source: img.src, observer: sessions[Number(enabled)].observer,
          surface: 'detail', variant: 'medium'
        });
        action.destroy();
      }
      return performance.now() - start;
    }
    for (let pair = -2; pair < 20; pair++) {
      for (const enabled of pair % 2 === 0 ? [false, true] : [true, false]) {
        const durationMs = batch(enabled);
        if (pair >= 0) samples.push({ pair, enabled, durationMs, perLifecycleMs: durationMs / perBatch });
      }
      // Flush outside the timed section; no scheduled timeout is part of the measurement.
      for (const callback of [...scheduled]) { scheduled.delete(callback); callback(); }
      const deadline = performance.now() + 5000;
      while (reported < (pair + 3) * perBatch || activeDeliveries > 0) {
        if (performance.now() > deadline) throw new Error('Controlled telemetry delivery did not settle');
        await new Promise(resolve => setTimeout(resolve, 1));
      }
    }
    sessions.forEach(session => session.dispose());
    await new Promise(resolve => setTimeout(resolve, 0));
    img.remove();
    function percentile(values: number[], fraction: number) {
      const sorted = [...values].sort((a, b) => a - b);
      return sorted[Math.ceil(sorted.length * fraction) - 1];
    }
    const summary = [false, true].map(enabled => {
      const values = samples.filter(s => s.enabled === enabled).map(s => s.perLifecycleMs);
      return { enabled, batchAverageP50Ms: percentile(values, .5), batchAverageP95Ms: percentile(values, .95) };
    });
    return { decoded, reported, failed, perBatch, samples, summary,
      medianDifferenceMs: summary[1].batchAverageP50Ms - summary[0].batchAverageP50Ms,
      runtime: navigator.userAgent,
      scope: 'Percentiles describe 20-lifecycle batch averages, not individual lifecycle tails. Cached decoded 512px synthetic image; DOM action plus production reporter overhead. Controlled transport/scheduler. Excludes decode, network, scheduling, rendering and physical-device latency.' };
  });
  expect(evidence.decoded).toBe(true);
  expect(evidence.reported).toBe(22 * evidence.perBatch);
  expect(evidence.failed).toBe(0);
  expect(evidence.samples).toHaveLength(40);
  expect(evidence.samples.every(s => Number.isFinite(s.durationMs) && s.durationMs >= 0)).toBe(true);
  const path = testInfo.outputPath('image-observer-overhead.json');
  await writeFile(path, JSON.stringify(evidence, null, 2));
  await testInfo.attach('image-observer-overhead', { path, contentType: 'application/json' });
});
