import type { ExportFileDelivery } from '$lib/ports/inventoryExport';

export class BrowserExportFileDelivery implements ExportFileDelivery {
  async save(file: Parameters<ExportFileDelivery['save']>[0], signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    const url = URL.createObjectURL(new Blob([file.content], { type: file.contentType }));
    const anchor = document.createElement('a');
    try {
      anchor.href = url;
      anchor.download = file.fileName;
      anchor.hidden = true;
      document.body.append(anchor);
      signal.throwIfAborted();
      anchor.click();
      // Allow the browser to accept the download before releasing its object URL.
      await new Promise<void>(resolve => setTimeout(resolve, 0));
    } finally {
      anchor.remove();
      URL.revokeObjectURL(url);
    }
  }
}
