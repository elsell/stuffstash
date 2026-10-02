import { afterEach, describe, expect, it } from 'vitest';
import { BrowserArchiveFileDelivery } from './BrowserArchiveFileDelivery';

const original = Object.getOwnPropertyDescriptor(window, 'showSaveFilePicker');
afterEach(() => { if (original) Object.defineProperty(window, 'showSaveFilePicker', original); else Reflect.deleteProperty(window, 'showSaveFilePicker'); });
describe('streaming archive file delivery', () => {
  it('chooses the destination before downloading and writes original bytes directly', async () => {
    const steps: string[] = [], bytes: number[] = [];
    const destination = new WritableStream<Uint8Array>({ write(chunk) { bytes.push(...chunk); }, close() { steps.push('closed'); } });
    Object.defineProperty(window, 'showSaveFilePicker', { configurable: true, value: async () => { steps.push('choose'); return { createWritable: async () => destination }; } });
    await new BrowserArchiveFileDelivery().save(async () => { steps.push('download'); return new ReadableStream({ start(controller) { controller.enqueue(new Uint8Array([80, 75, 0, 255])); controller.close(); } }); }, new AbortController().signal);
    expect(steps).toEqual(['choose', 'download', 'closed']); expect(bytes).toEqual([80, 75, 0, 255]);
  });
  it('aborts the destination rather than saving an API error response', async () => {
    let aborted = false;
    Object.defineProperty(window, 'showSaveFilePicker', { configurable: true, value: async () => ({ createWritable: async () => new WritableStream({ abort() { aborted = true; } }) }) });
    await expect(new BrowserArchiveFileDelivery().save(async () => { throw new Error('Forbidden'); }, new AbortController().signal)).rejects.toThrow('Forbidden');
    expect(aborted).toBe(true);
  });
  it('cancels a fallback transfer when the task leaves its scope', async () => {
    Reflect.deleteProperty(window, 'showSaveFilePicker');
    const controller = new AbortController(); let cancelled = false;
    const pending = new BrowserArchiveFileDelivery().save(async () => new ReadableStream({ cancel() { cancelled = true; } }), controller.signal);
    await Promise.resolve(); controller.abort();
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' }); expect(cancelled).toBe(true);
  });
});
