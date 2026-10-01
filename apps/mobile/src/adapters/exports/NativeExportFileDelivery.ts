import type { InventoryExportFile, InventoryExportFileDelivery, InventoryExportObserver } from '../../application/exports/InventoryExport';
import { assertReadActive } from '../../application/shared/ReadRequest';

export interface ExportTemporaryFiles {
  sweep(): Promise<void>;
  write(file: InventoryExportFile): Promise<{ uri: string; finish(retain: boolean): Promise<void> }>;
}
export interface ExportShareSheet {
  isAvailable(): Promise<boolean>;
  open(uri: string, format: InventoryExportFile['format']): Promise<void>;
}

export class NativeExportFileDelivery implements InventoryExportFileDelivery {
  constructor(private readonly files: ExportTemporaryFiles, private readonly sheet: ExportShareSheet, private readonly platform: 'ios' | 'android', private readonly observer: InventoryExportObserver) {}
  async share(file: InventoryExportFile, signal: AbortSignal): Promise<void> {
    assertReadActive(signal);
    if (!await this.sheet.isAvailable()) throw new Error('File sharing is unavailable on this device.');
    assertReadActive(signal);
    await this.files.sweep();
    assertReadActive(signal);
    const temporary = await this.files.write(file);
    let handedOff = false;
    const finish = async (retain: boolean) => {
      try { await temporary.finish(retain); }
      catch { this.observer.record({ name: 'inventory_export.cleanup_failed' }); }
    };
    try {
      assertReadActive(signal);
      const sharing = this.sheet.open(temporary.uri, file.format);
      handedOff = true;
      const completion = sharing.then(async () => { await finish(this.platform === 'android'); }, async error => { await finish(false); throw error; });
      // The native presentation continues owning its file even if this screen leaves.
      await whileActive(completion, signal);
    } finally {
      if (!handedOff) await finish(false);
    }
  }
}

function whileActive(operation: Promise<void>, signal: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const abort = () => { try { assertReadActive(signal); } catch (error) { reject(error); } };
    signal.addEventListener('abort', abort, { once: true });
    operation.then(resolve, reject).finally(() => signal.removeEventListener('abort', abort));
    if (signal.aborted) abort();
  });
}
