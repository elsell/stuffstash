import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const styles = readFileSync('src/styles.css', 'utf8');

describe('responsive workspace row styles', () => {
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
});
