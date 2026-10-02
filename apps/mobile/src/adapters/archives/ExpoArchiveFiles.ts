import { Directory, File, Paths } from 'expo-file-system';
import * as Crypto from 'expo-crypto';
import * as Sharing from 'expo-sharing';
import { Platform } from 'react-native';
import { assertReadActive } from '../../application/shared/ReadRequest';
import type { ArchiveFiles, PickedArchive } from '../../application/archives/InventoryArchive';
import type { InventoryExportObserver } from '../../application/exports/InventoryExport';
import { t } from '../../presentation/localization';

const retentionMs = 24 * 60 * 60 * 1000;
const root = () => new Directory(Paths.cache, 'inventory-archives');
export class ExpoArchiveFiles implements ArchiveFiles {
  private readonly active = new Set<string>();
  constructor(private readonly observer: InventoryExportObserver) {}
  async pick(signal: AbortSignal): Promise<PickedArchive | undefined> {
    assertReadActive(signal);
    let result: File | File[];
    try { result = await File.pickFileAsync(undefined, 'application/zip'); }
    catch (error) { if ((error as { code?: string }).code === 'ERR_FILE_PICKING_CANCELLED') return; throw error; }
    const file = Array.isArray(result) ? result[0] : result;
    if (!file) return;
    const dispose = () => {
      // The pinned iOS picker uses asCopy:true. Its returned app-owned copy
      // can live in UIKit temporary/Inbox storage, not just Paths.cache.
      // Never delete an Android document-provider source.
      if (Platform.OS === 'ios' && file.uri.startsWith('file://')) {
        try { if (file.exists) file.delete(); } catch { this.cleanupFailed(); }
      }
    };
    try { assertReadActive(signal); return { name: file.name, uri: file.uri, dispose }; }
    catch (error) { dispose(); throw error; }
  }
  async share(load: () => Promise<ReadableStream<Uint8Array>>, signal: AbortSignal): Promise<void> {
    assertReadActive(signal);
    if (!await Sharing.isAvailableAsync()) throw new Error(t('recovery.sharingUnavailable'));
    assertReadActive(signal);
    this.sweep();
    const directory = new Directory(root(), `${Date.now()}-${Crypto.randomUUID()}`);
    directory.create({ intermediates: true }); this.active.add(directory.uri);
    const file = new File(directory, 'stuff-stash-inventory.zip');
    let handedOff = false;
    try {
      file.create();
      const stream = await load();
      await stream.pipeTo(file.writableStream(), { signal });
      assertReadActive(signal);
      const sharing = Sharing.shareAsync(file.uri, { dialogTitle: t('archive.export'), mimeType: 'application/zip', UTI: 'public.zip-archive' });
      handedOff = true;
      // Once presented, the native share sheet owns the file until it completes.
      const completion = sharing.then(() => this.finish(directory, Platform.OS === 'android'), error => { this.finish(directory, false); throw error; });
      await untilAborted(completion, signal);
    } finally { if (!handedOff) this.finish(directory, false); }
  }
  sweep(): void {
    const directory = root(); if (!directory.exists) return;
    for (const entry of directory.list()) {
      if (!(entry instanceof Directory) || this.active.has(entry.uri)) continue;
      const created = Number(entry.name.split('-')[0]);
      if (Number.isFinite(created) && created > 0 && Date.now() - created >= retentionMs) this.remove(entry);
    }
  }
  private finish(directory: Directory, retain: boolean) {
    this.active.delete(directory.uri);
    if (retain) setTimeout(() => this.remove(directory), retentionMs);
    else this.remove(directory);
  }
  private remove(directory: Directory) { try { if (directory.exists) directory.delete(); } catch { this.cleanupFailed(); } }
  private cleanupFailed() { this.observer.record({ name: 'inventory_export.cleanup_failed' }); }
}
function untilAborted(operation: Promise<void>, signal: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const abort = () => { try { assertReadActive(signal); } catch (error) { reject(error); } };
    signal.addEventListener('abort', abort, { once: true });
    operation.then(resolve, reject).finally(() => signal.removeEventListener('abort', abort));
    if (signal.aborted) abort();
  });
}
