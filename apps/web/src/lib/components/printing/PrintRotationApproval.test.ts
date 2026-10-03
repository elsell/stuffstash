import { afterEach, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import PrintRotationApproval from './PrintRotationApproval.svelte';
import { PairingFailure } from '$lib/domain/printPairing';
class RotationRepository {
    authorized = true;
    generation = 1;
    approved = false;
    async inventories() { return this.authorized ? [{ tenantId: 't', inventoryId: 'i', name: 'Home', tenantName: 'Family' }] : []; }
    async connector(scope: {
        tenantId: string;
        inventoryId: string;
    }, id: string) { if (!this.authorized || scope.tenantId !== 't' || scope.inventoryId !== 'i' || id !== 'c')
        throw new PairingFailure('denied'); return { id, name: 'Garage computer', generation: this.generation }; }
    async review(id: string, scope: {
        tenantId: string;
        inventoryId: string;
    }, code: string) { await this.connector(scope, 'c'); if (id !== 'pair' || code !== 'ABCD1234')
        throw new PairingFailure('invalid'); return { id, name: 'Replacement', fingerprint: 'public', candidates: [], rotation: true }; }
    async rotate(id: string, scope: {
        tenantId: string;
        inventoryId: string;
    }, code: string, connectorId: string, generation: number) { await this.review(id, scope, code); const c = await this.connector(scope, connectorId); if (c.generation !== generation)
        throw new PairingFailure('invalid'); this.approved = true; }
}
let component: ReturnType<typeof mount> | undefined;
afterEach(() => { if (component)
    unmount(component); component = undefined; document.body.innerHTML = ''; });
async function settle() { for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    await tick();
} }
async function start(repo: RotationRepository, target = { tenantId: 't', inventoryId: 'i', connectorId: 'c' }) { component = mount(PrintRotationApproval, { target: document.body, props: { pairingId: 'pair', repository: repo, target, onSignIn: async () => { } } }); await settle(); }
async function review() { const field = document.querySelector('input')!; field.value = 'ABCD1234'; field.dispatchEvent(new Event('input', { bubbles: true })); await settle(); document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); await settle(); }
it('requires explicit approval of the exact reviewed connector generation', async () => { const repo = new RotationRepository(); await start(repo); await review(); expect(repo.approved).toBe(false); const button = Array.from(document.querySelectorAll('button')).find(b => b.textContent?.includes('Replace credential'))!; expect(button).toBeTruthy(); repo.generation++; button.click(); await settle(); expect(repo.approved).toBe(false); expect(document.querySelector('[role="alert"]')).toBeTruthy(); });
it('does not fall back to another inventory when the target is inaccessible', async () => { const repo = new RotationRepository(); await start(repo, { tenantId: 'other', inventoryId: 'i', connectorId: 'c' }); expect(document.querySelector('input')).toBeNull(); expect(document.querySelector('[role="alert"]')).toBeTruthy(); expect(repo.approved).toBe(false); });
it('approves the existing connector without registering a printer', async () => { const repo = new RotationRepository(); await start(repo); await review(); Array.from(document.querySelectorAll('button')).find(b => b.textContent?.includes('Replace credential'))!.click(); await settle(); expect(repo.approved).toBe(true); expect(document.body.textContent).toContain('approved'); });
