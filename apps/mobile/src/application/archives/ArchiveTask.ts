import { assertReadActive } from '../shared/ReadRequest';
import type { ArchiveScope, InventoryArchiveWorkspace, PickedArchive } from './InventoryArchive';

/** Owns a screen's local transfers; durable server jobs outlive this task. */
export class ArchiveTask {
  private readonly lifetime = new AbortController();
  private source?: PickedArchive;
  private uploadKey?: string;
  private creation?: { photos: boolean; otherFiles: boolean; key: string };
  private busy = false;
  constructor(private readonly workspace: InventoryArchiveWorkspace, readonly scope: ArchiveScope) {}
  get signal() { return this.lifetime.signal; }
  get fileName() { return this.source?.name; }
  close() {
    this.lifetime.abort();
    // An in-flight upload may still be reading its file while native cancellation
    // settles. Its finally block owns disposal in that case.
    if (!this.busy) this.discardSource();
  }
  async pick() {
    return this.exclusive(async () => {
      const picked = await this.workspace.files.pick(this.signal);
      if (this.signal.aborted) { picked?.dispose(); assertReadActive(this.signal); }
      if (picked) { this.discardSource(); this.source = picked; this.uploadKey = this.workspace.newRequestKey(); }
      return this.fileName;
    });
  }
  async upload() {
    return this.exclusive(async () => {
      if (!this.source || !this.uploadKey) throw new Error('No archive selected');
      const job = await this.workspace.repository.upload(this.scope.tenantId, this.uploadKey, this.source.uri, this.signal);
      this.discardSource();
      return job;
    });
  }
  async create(selection: { photos: boolean; otherFiles: boolean }) {
    return this.exclusive(async () => {
      const { tenantId, inventoryId } = this.scope;
      if (!inventoryId) throw new Error('Export requires an inventory');
      if (!this.creation || this.creation.photos !== selection.photos || this.creation.otherFiles !== selection.otherFiles) {
        this.creation = { ...selection, key: this.workspace.newRequestKey() };
      }
      const job = await this.workspace.repository.create({ tenantId, inventoryId }, this.creation.key, selection, this.signal);
      this.creation = undefined;
      return job;
    });
  }
  private discardSource() { this.source?.dispose(); this.source = undefined; this.uploadKey = undefined; }
  private async exclusive<T>(operation: () => Promise<T>): Promise<T> {
    assertReadActive(this.signal);
    if (this.busy) throw new Error('Archive task already running');
    this.busy = true;
    try { const result = await operation(); assertReadActive(this.signal); return result; }
    finally { this.busy = false; if (this.signal.aborted) this.discardSource(); }
  }
}
