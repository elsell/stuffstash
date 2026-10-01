import { describe, expect, it } from 'vitest';
import { NativeExportFileDelivery, type ExportTemporaryFiles, type ExportShareSheet } from './NativeExportFileDelivery';

class TemporaryFiles implements ExportTemporaryFiles {
  written = false; removed = false; retained = false; sweeps = 0;
  async sweep() { this.sweeps++; }
  async write() { this.written = true; return { uri: 'file:///cache/export.json', finish: async (retain: boolean) => { this.removed = !retain; this.retained = retain; } }; }
}
class ShareSheet implements ExportShareSheet {
  available = true; entered = false;
  finish!: () => void;
  fail!: (error: Error) => void;
  async isAvailable() { return this.available; }
  async open() { this.entered = true; await new Promise<void>((resolve, reject) => { this.finish = resolve; this.fail = reject; }); }
}
async function entered(sheet: ShareSheet) { for (let i = 0; i < 10 && !sheet.entered; i++) await Promise.resolve(); expect(sheet.entered).toBe(true); }
const observer = { record() {} };

describe('native export handoff lifetime', () => {
  it.each(['ios', 'android'] as const)('keeps the file until the %s handoff completes and applies its retention policy', async platform => {
    const files = new TemporaryFiles(); const sheet = new ShareSheet();
    const pending = new NativeExportFileDelivery(files, sheet, platform, observer).share({ content: '{}', format: 'json' }, new AbortController().signal);
    await entered(sheet); expect(files.removed).toBe(false);
    sheet.finish(); await pending;
    expect(files.removed).toBe(platform === 'ios'); expect(files.retained).toBe(platform === 'android');
  });
  it('cancels ownership without deleting a file still used by the share sheet', async () => {
    const files = new TemporaryFiles(); const sheet = new ShareSheet(); const cancellation = new AbortController();
    const pending = new NativeExportFileDelivery(files, sheet, 'ios', observer).share({ content: '{}', format: 'json' }, cancellation.signal);
    await entered(sheet); cancellation.abort();
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' }); expect(files.removed).toBe(false);
    sheet.finish(); for (let i = 0; i < 10; i++) await Promise.resolve(); expect(files.removed).toBe(true);
  });
  it('removes a failed handoff and never writes when sharing is unavailable', async () => {
    const files = new TemporaryFiles(); const sheet = new ShareSheet();
    const delivery = new NativeExportFileDelivery(files, sheet, 'android', observer);
    const pending = delivery.share({ content: '{}', format: 'json' }, new AbortController().signal);
    await entered(sheet); sheet.fail(new Error('Sharing failed'));
    await expect(pending).rejects.toThrow('Sharing failed'); expect(files.removed).toBe(true);
    const unavailableFiles = new TemporaryFiles(); sheet.available = false;
    await expect(new NativeExportFileDelivery(unavailableFiles, sheet, 'ios', observer).share({ content: '{}', format: 'json' }, new AbortController().signal)).rejects.toThrow('unavailable');
    expect(unavailableFiles.written).toBe(false);
  });
});
