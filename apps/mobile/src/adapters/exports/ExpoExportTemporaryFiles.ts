import { Directory, File, Paths } from 'expo-file-system';
import * as Crypto from 'expo-crypto';
import type { InventoryExportFile, InventoryExportObserver } from '../../application/exports/InventoryExport';
import type { ExportTemporaryFiles } from './NativeExportFileDelivery';

const retentionMs = 24 * 60 * 60 * 1000;
const root = () => new Directory(Paths.cache, 'inventory-exports');

export class ExpoExportTemporaryFiles implements ExportTemporaryFiles {
  private readonly active = new Set<string>();
  constructor(private readonly observer: InventoryExportObserver) {}
  async sweep(): Promise<void> {
    const directory = root();
    if (!directory.exists) return;
    for (const entry of directory.list()) {
      if (!(entry instanceof Directory) || this.active.has(entry.uri)) continue;
      const createdAt = Number(entry.name.split('-')[0]);
      if (Number.isFinite(createdAt) && createdAt > 0 && Date.now() - createdAt >= retentionMs) this.remove(entry);
    }
  }
  async write(file: InventoryExportFile) {
    const directory = new Directory(root(), `${Date.now()}-${Crypto.randomUUID()}`);
    directory.create({ intermediates: true });
    this.active.add(directory.uri);
    const target = new File(directory, `stuff-stash-inventory.${file.format}`);
    try { target.write(file.content); }
    catch (error) { this.active.delete(directory.uri); this.remove(directory); throw error; }
    return { uri: target.uri, finish: async (retain: boolean) => {
      this.active.delete(directory.uri);
      if (retain) {
        setTimeout(() => this.remove(directory), retentionMs);
      } else this.remove(directory);
    } };
  }
  private remove(directory: Directory) {
    try { if (directory.exists) directory.delete(); }
    catch { this.observer.record({ name: 'inventory_export.cleanup_failed' }); }
  }
}
