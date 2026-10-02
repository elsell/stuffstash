import type { ArchiveFileDelivery } from '$lib/ports/inventoryArchive';

const blobDownloadLimit = 128 * 1024 * 1024;
type SavePickerWindow = Window & { showSaveFilePicker?: (options: { suggestedName: string; types: { description: string; accept: Record<string, string[]> }[] }) => Promise<FileSystemFileHandle> };

export class BrowserArchiveFileDelivery implements ArchiveFileDelivery {
  async save(load: () => Promise<ReadableStream<Uint8Array>>, signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    const picker = window as SavePickerWindow;
    if (picker.showSaveFilePicker) {
      // Acquire the destination during the click gesture, before any network await.
      const handle = await picker.showSaveFilePicker({ suggestedName: 'stuff-stash-inventory.zip', types: [{ description: 'Stuff Stash archive', accept: { 'application/zip': ['.zip'] } }] });
      signal.throwIfAborted();
      const output = await handle.createWritable();
      try {
        const content = await load();
        await content.pipeTo(output, { signal });
      } catch (error) { try { await output.abort(error); } catch { /* pipeTo may already have aborted it. */ } throw error; }
      return;
    }
    const content = await load();
    const reader = content.getReader();
    const abort = () => { void reader.cancel(signal.reason); };
    signal.addEventListener('abort', abort, { once: true });
    const chunks: Uint8Array<ArrayBuffer>[] = [];
    let size = 0;
    try {
      signal.throwIfAborted();
      while (true) {
        const { done, value } = await reader.read();
        signal.throwIfAborted();
        if (done) break;
        size += value.byteLength;
        if (size > blobDownloadLimit) throw Object.assign(new Error('Streaming file save is required for this archive.'), { code: 'archive_streaming_save_required' });
        chunks.push(new Uint8Array(value));
      }
    } catch (error) { await reader.cancel(error); throw error; }
    finally { signal.removeEventListener('abort', abort); reader.releaseLock(); }
    const url = URL.createObjectURL(new Blob(chunks, { type: 'application/zip' }));
    const anchor = document.createElement('a');
    try {
      anchor.href = url; anchor.download = 'stuff-stash-inventory.zip'; anchor.hidden = true;
      document.body.append(anchor); signal.throwIfAborted(); anchor.click();
      await new Promise<void>(resolve => setTimeout(resolve, 0));
    } finally { anchor.remove(); URL.revokeObjectURL(url); }
  }
}
