import React, { useLayoutEffect, useSyncExternalStore } from 'react';
import { getScreenFocused, setScreenFocused, subscribeScreenFocus } from '../../test-support/navigation';
import { describe, expect, it } from 'vitest';
import { ExportInventoryCommand, type InventoryExportFile } from '../../application/exports/InventoryExport';
import { MobileRenderHarness } from '../../test-support/render';
import { InventoryExportAction } from './InventoryExportAction';

async function choose(h: MobileRenderHarness, label: string) {
  await h.press(h.byLabel('Export inventory'));
  await h.press(h.allByType('Pressable').find(node => node.queryAll(child => child.type === 'Text' && child.children.includes(label)).length > 0));
}
describe('inventory export settings action', () => {
  it.each(['scope', 'blur', 'unmount'] as const)('retires pending content during the %s commit before it can share', async transition => {
    const h = new MobileRenderHarness(); const delivered: InventoryExportFile[] = [];
    let finish!: (value: string) => void; let signal!: AbortSignal; let abortedAtCommit = false;
    const command = new ExportInventoryCommand({ download: (_scope, _format, requestSignal) => {
      signal = requestSignal; return new Promise(resolve => { finish = resolve; });
    } }, { async share(file) { delivered.push(file); } });
    function Host({ inventory = 'main', visible = true }: { inventory?: string; visible?: boolean }) {
      const focused = useSyncExternalStore(subscribeScreenFocus, getScreenFocused);
      useLayoutEffect(() => {
        if (signal && (inventory !== 'main' || !visible || !focused)) {
          abortedAtCommit = signal.aborted; finish('retired inventory content');
        }
      }, [inventory, visible, focused]);
      return visible ? <InventoryExportAction command={command} scope={{ tenantId: 'home', inventoryId: inventory }} /> : null;
    }
    await h.render(<Host />); await choose(h, 'JSON — complete inventory data');
    try {
      if (transition === 'scope') await h.render(<Host inventory="other" />);
      else if (transition === 'blur') await h.run(() => setScreenFocused(false));
      else await h.render(<Host visible={false} />);
      expect(abortedAtCommit).toBe(true); expect(delivered).toEqual([]);
    } finally { await h.unmount(); setScreenFocused(true); }
  });
  it('prevents duplicate requests, cancels in place, and does not share late content', async () => {
    const h = new MobileRenderHarness(); let reads = 0; let finish!: (value: string) => void;
    const delivered: InventoryExportFile[] = [];
    const command = new ExportInventoryCommand({ download: () => { reads++; return new Promise(resolve => { finish = resolve; }); } }, { async share(file) { delivered.push(file); } });
    await h.render(<InventoryExportAction command={command} scope={{ tenantId: 'home', inventoryId: 'main' }} />);
    await choose(h, 'JSON — complete inventory data');
    expect(reads).toBe(1); expect(h.byText('Preparing export…')).toBeDefined();
    expect(h.byLabel('Export inventory')?.props.disabled).toBe(true);
    await h.press(h.byLabel('Cancel export'));
    await h.run(async () => { finish('private content'); });
    expect(delivered).toEqual([]);
    expect(h.byLabel('Export inventory')?.props.disabled).toBe(false);
    await h.unmount();
  });
  it('retains a retry path after failure and uses the last format', async () => {
    const h = new MobileRenderHarness(); let attempts = 0; const formats: string[] = [];
    const command = new ExportInventoryCommand({ async download(_scope, format) { formats.push(format); if (++attempts === 1) throw new Error('Network unavailable'); return 'id,title'; } }, { async share() {} });
    await h.render(<InventoryExportAction command={command} scope={{ tenantId: 'home', inventoryId: 'main' }} />);
    await choose(h, 'CSV — spreadsheet rows');
    expect(h.byText('Could not export this inventory. Try again.')).toBeDefined();
    await h.press(h.byLabel('Retry export'));
    expect(formats).toEqual(['csv', 'csv']);
    expect(h.byText('Could not export this inventory. Try again.')).toBeUndefined();
    await h.unmount();
  });
});
