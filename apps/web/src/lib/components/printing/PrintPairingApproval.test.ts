import { afterEach, expect, it } from 'vitest';
import { mount, tick, unmount, flushSync } from 'svelte';
import PrintPairingApproval from './PrintPairingApproval.svelte';
import { PairingFailure, type PairingInventory } from '$lib/domain/printPairing';
import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
class ApprovalRepository implements PrintPairingRepository {
    authorized = true;
    approved = false;
    scope: PairingInventory = { tenantId: 'tenant', inventoryId: 'inventory', name: 'Home', tenantName: 'Family' };
    async inventories() { return this.authorized ? [this.scope] : []; }
    async review(_id: string, scope: PairingInventory, code: string) {
        if (!this.authorized)
            throw new PairingFailure('denied');
        if (code !== 'ABCD1234' || scope.inventoryId !== this.scope.inventoryId)
            throw new PairingFailure('invalid');
        return { id: 'pair', name: 'Garage computer', fingerprint: 'sha256:public-fingerprint', candidates: [{ id: 'candidate', name: 'Brother QL-800', adapterId: 'brother' }] };
    }
    async setup() { return { printers: [{ id: 'printer', name: 'Garage printer', adapterId: 'brother', mediaName: '29 × 90 mm' }], media: [] }; }
    async createPrinter(): Promise<string> { throw new Error('fixture has existing destination'); }
    async approve(id: string, scope: PairingInventory, code: string, bindings: {
        candidateId: string;
        printerId: string;
    }[]) {
        await this.review(id, scope, code);
        if (bindings.length !== 1 || bindings[0].candidateId !== 'candidate' || bindings[0].printerId !== 'printer')
            throw new PairingFailure('invalid');
        this.approved = true;
    }
}
let component: ReturnType<typeof mount> | undefined;
afterEach(() => { if (component)
    unmount(component); component = undefined; document.body.innerHTML = ''; });
async function settle() { for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await tick();
} }
async function start(repo: ApprovalRepository) { component = mount(PrintPairingApproval, { target: document.body, props: { pairingId: 'pair', repository: repo, onSignIn: async () => { } } }); await settle(); }
function control<T extends HTMLInputElement | HTMLSelectElement>(id: string): T { return document.getElementById(id) as T; }
async function input(id: string, value: string) {
 const node=document.getElementById(id)!;
 if(node instanceof HTMLInputElement){node.value=value;node.dispatchEvent(new Event('input',{bubbles:true}));}
 else{
  const label=value==='0'?'Family / Home':'Garage printer — 29 × 90 mm';
  node.focus();node.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();
  const option=()=>Array.from(document.querySelectorAll<HTMLElement>('[role="option"]')).find(item=>item.textContent?.trim()===label);
  await expect.poll(()=>Boolean(option())).toBe(true);
  option()!.dispatchEvent(new PointerEvent('pointerup',{pointerType:'mouse',bubbles:true,cancelable:true}));flushSync();
 }
 await settle();
}
async function submit(index: number) { document.querySelectorAll('form')[index].dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); await settle(); }
it('keeps pairing identity hidden until valid code; approval requires explicit compatible destination', async () => {
    const repo = new ApprovalRepository();
    await start(repo);
    await input('pair-inventory', '0');
    await input('pair-code', 'BADCODE1');
    await submit(0);
    expect(document.body.textContent).not.toContain('Garage computer');
    expect(repo.approved).toBe(false);
    await input('pair-code', 'ABCD1234');
    await submit(0);
    expect(document.body.textContent).toContain('Garage computer');
    expect(repo.approved).toBe(false);
    expect(document.getElementById('destination-0')?.textContent).toContain('Do not connect this printer');
    await input('destination-0', 'printer');
    await submit(1);
    expect(repo.approved).toBe(true);
    expect(document.body.textContent).toContain('Connection approved');
});
it('clears a reviewed identity when code changes and handles permission loss safely', async () => {
    const repo = new ApprovalRepository();
    await start(repo);
    await input('pair-inventory', '0');
    await input('pair-code', 'ABCD1234');
    await submit(0);
    await input('pair-code', 'BADCODE1');
    expect(document.body.textContent).not.toContain('Garage computer');
    await input('pair-code', 'ABCD1234');
    await submit(0);
    await input('destination-0', 'printer');
    repo.authorized = false;
    await submit(1);
    expect(repo.approved).toBe(false);
    expect(document.body.textContent).toContain('no longer have permission');
});
it('offers no approval controls without a configurable inventory', async () => {
    const repo = new ApprovalRepository();
    repo.authorized = false;
    await start(repo);
    expect(document.body.textContent).toContain('do not have permission');
    expect(document.querySelector('form')).toBeNull();
});
