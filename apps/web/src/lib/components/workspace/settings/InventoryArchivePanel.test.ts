import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import InventoryArchivePanel from './InventoryArchivePanel.svelte';
import type { ArchiveJob, InventoryArchiveRepository } from '$lib/ports/inventoryArchive';

let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const job: ArchiveJob = { id: 'job', kind: 'export', state: 'queued', phase: 'execution', createdAt: '2026-10-02T12:00:00Z', expiresAt: '2099-10-03T12:00:00Z', photos: true, otherFiles: true };
class Repository implements InventoryArchiveRepository {
  jobs: ArchiveJob[] = []; keys: string[] = []; fail = false; approvals: string[] = [];
  async list() { return { jobs: this.jobs }; }
  async get(_scope: unknown, id: string) { return this.jobs.find(j => j.id === id)!; }
  async create(_scope: unknown, key: string) { this.keys.push(key); if (this.fail) throw new Error('Connection lost'); this.jobs = [job]; return job; }
  async upload() { return job; }
  async preview() { return { inventoryName: 'Garage', assets: 3, tags: 2, customAssetTypes: 0, customFields: 0, photos: 1, otherFiles: 0, omittedAttachments: 2, keyRemappings: [] }; }
  async approve(_tenant: string, _id: string, name: string) { this.approvals.push(name); this.jobs = [{ ...job, kind: 'restore', state: 'running' }]; return this.jobs[0]!; }
  async cancel() { return { ...job, state: 'cancelled' as const }; }
  async retry() { return job; }
  async download() { return new ReadableStream<Uint8Array>(); }
}
async function settle() { for (let n = 0; n < 8; n++) { await Promise.resolve(); await tick(); } }
function button(text: string) { const element = [...document.querySelectorAll('button')].find(b => b.textContent?.trim() === text); if (!element) throw new Error(`Missing ${text}`); return element; }
function render(repository: Repository, inventoryId?: string) {
  component = mount(InventoryArchivePanel, { target: document.body, props: { workspace: { repository, files: { async save() {} } }, scope: { tenantId: 'home', inventoryId } } });
}
describe('archive tasks', () => {
  it('keeps published restores pending without offering cancellation or navigation', async () => {
    const repository = new Repository();
    repository.jobs = [{ ...job, kind: 'restore', state: 'queued', phase: 'finalization', destinationInventoryId: 'restored' }];
    render(repository); await settle();
    expect(document.body.textContent).toContain('Restoring');
    expect([...document.querySelectorAll('button')].some(b => b.textContent?.trim() === 'Cancel job')).toBe(false);
    expect(document.querySelector('a')).toBeNull();
  });
  it('reuses the creation key after a lost response and never cancels the job on exit', async () => {
    const repository = new Repository(); repository.fail = true; render(repository, 'main'); await settle();
    button('Create archive').click(); await settle(); expect(document.querySelector('[role="alert"]')).not.toBeNull();
    repository.fail = false; button('Create archive').click(); await settle();
    expect(repository.keys).toHaveLength(2); expect(repository.keys[0]).toBe(repository.keys[1]);
    expect(document.body.textContent).toContain('Queued');
    expect([...document.querySelectorAll('button')].some(b => b.textContent === 'Download')).toBe(false);
  });
  it('does not let an older list response erase a newly created job', async () => {
    class DeferredRepository extends Repository {
      finish!: (value: {jobs: ArchiveJob[]}) => void;
      override list(): Promise<{jobs: ArchiveJob[]}> { return new Promise(resolve => { this.finish = resolve; }); }
    }
    const repository = new DeferredRepository(); render(repository, 'main'); await settle();
    button('Create archive').click(); await settle();
    repository.finish({ jobs: [] }); await settle();
    expect(document.body.textContent).toContain('Queued');
  });
  it('requires review and a name before restoring a new inventory', async () => {
    const repository = new Repository(); repository.jobs = [{ ...job, kind: 'restore', state: 'awaiting_approval', phase: 'validation' }]; render(repository); await settle();
    button('Review restore').click(); await settle();
    expect(document.body.textContent).toContain('2 attachments omitted');
    const name = document.querySelector<HTMLInputElement>('input[name="inventoryName"]')!;
    expect(name.value).toBe('Garage');
    name.value = ''; name.dispatchEvent(new Event('input', { bubbles: true })); await settle(); expect(button('Restore inventory').disabled).toBe(true);
    name.value = 'Imported garage'; name.dispatchEvent(new Event('input', { bubbles: true })); await settle();
    button('Restore inventory').click(); await settle(); expect(repository.approvals).toEqual(['Imported garage']);
    expect(document.body.textContent).not.toContain('Review restore');
  });
});
