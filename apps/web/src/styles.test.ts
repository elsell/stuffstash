import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

const styles = readFileSync('src/styles.css', 'utf8');

describe('responsive workspace row styles', () => {
  it('keeps workspace button targets at least 44 CSS pixels high', () => {
    const workspaceButtonRule = styles.match(/^\.workspace-main \[data-slot="button"\]\s*\{([^}]*)\}/m)?.[1] ?? '';
    expect(workspaceButtonRule).toContain('min-height: 44px');
  });

  it('keeps asset content in the shrinkable column before row actions', () => {
    const assetRowRules = [...styles.matchAll(/^\s*\.asset-row\s*\{([^}]*)\}/gm)];
    const narrowAssetRowRule = assetRowRules.at(-1)?.[1] ?? '';
    expect(narrowAssetRowRule).toContain('grid-template-columns: minmax(0, 1fr) auto');
  });

  it('lets row links grow around thumbnails and wrapped titles', () => {
    const assetRowOpenRule = styles.match(/^\.asset-row-open\s*\{([^}]*)\}/m)?.[1] ?? '';
    expect(assetRowOpenRule).toContain('height: auto');
    expect(assetRowOpenRule).toContain('min-height: 48px');
  });

  it('gives every recent-card thumbnail the same full-width grid track', () => {
    const recentCardOpenRule = styles.match(/^\.recent-card-open\s*\{([^}]*)\}/m)?.[1] ?? '';
    expect(recentCardOpenRule).toContain('grid-template-columns: minmax(0, 1fr)');
  });
});
