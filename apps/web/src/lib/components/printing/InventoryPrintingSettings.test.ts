import { afterEach, expect, it } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import InventoryPrintingSettings from './InventoryPrintingSettings.svelte';
import { FakePrintingRepository } from '$lib/fakes/printingRepository';
let component: ReturnType<typeof mount> | undefined;
afterEach(() => { if (component)
    unmount(component); component = undefined; document.body.innerHTML = ''; });
async function settle() { for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    await tick();
} }
async function start(repo: FakePrintingRepository, canConfigure = true) { component = mount(InventoryPrintingSettings, { target: document.body, props: { scope: repo.scope, repository: repo, canConfigure, canPrint: canConfigure } }); await settle(); }
function button(text: string) { return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent?.trim() === text)!; }
it('shows readiness separately from heartbeat and denies viewer edits', async () => { const repo = new FakePrintingRepository(); repo.canConfigure = false; await start(repo, false); expect(document.body.textContent).toContain('Garage Brother'); expect(document.body.textContent).toContain('Garage computer'); expect((document.getElementById('default-auto-print') as HTMLInputElement).disabled).toBe(true); expect(button('Save defaults')).toBeUndefined(); });
it('keeps unsaved choice on status refresh and permits an offline default', async () => { const repo = new FakePrintingRepository(); await start(repo); const check = document.getElementById('default-auto-print') as HTMLInputElement; check.click(); await settle(); button('Refresh status').click(); await settle(); expect(check.checked).toBe(true); document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); await settle(); expect(repo.defaults.printOnCreateDefault).toBe(true); });
it('does not overwrite another editor and keeps stale draft visible', async () => { const repo = new FakePrintingRepository(); await start(repo); repo.defaults.revision++; document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); await settle(); expect(document.querySelector('[role="alert"]')).not.toBeNull(); expect((document.getElementById('default-auto-print') as HTMLInputElement).disabled).toBe(true); expect(repo.defaults.revision).toBe(1); });
