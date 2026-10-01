import type { InventoryConversationConnection, InventoryConversationEvent, InventoryConversationPlan, InventoryConversationScope, InventoryConversationTransport, ConversationAssetReference } from '$lib/ports/inventoryConversation';
export interface InventoryConversationState {
  messages: { role: 'user' | 'assistant'; text: string; assets: ConversationAssetReference[] }[];
  busy: boolean; plan: InventoryConversationPlan | null; error: string; uncertain: boolean;
}
export class InventoryConversation {
  state: InventoryConversationState = { messages: [], busy: false, plan: null, error: '', uncertain: false };
  private connection?: InventoryConversationConnection;
  private controller?: AbortController;
  private generation = 0;
  private disposed = false;
  private approving = false;
  constructor(private transport: InventoryConversationTransport, private scope: InventoryConversationScope,
    private changed: () => void, private refresh: () => boolean | void | Promise<boolean | void>, private authenticationLost: () => void) {}
  async send(raw: string): Promise<void> {
    const text = raw.trim();
    if (this.disposed || this.state.busy || this.state.plan || this.state.uncertain || !text || [...text].length > 8000) return;
    this.state = { ...this.state, busy: true, error: '', messages: [...this.state.messages, { role: 'user', text, assets: [] }] }; this.changed();
    const generation = this.generation;
    try {
      if (!this.connection) {
        const controller = new AbortController(); this.controller = controller;
        const connection = await this.transport.connect(this.scope, controller.signal,
          (event) => { if (generation === this.generation && !this.disposed) this.receive(event); },
          (kind) => { if (generation === this.generation && !this.disposed) this.fail(kind); });
        if (controller.signal.aborted || generation !== this.generation || this.disposed) { connection.close(); return; }
        this.connection = connection;
      }
      this.connection.send(text);
    } catch { if (generation === this.generation && !this.disposed) this.fail('unavailable'); }
  }
  decide(approve: boolean): void {
    if (this.disposed || this.state.busy || !this.state.plan || !this.connection) return;
    const id = this.state.plan.id; this.approving = approve;
    this.state = { ...this.state, busy: true }; this.changed();
    try { this.connection.decide(id, approve); } catch { this.fail('unavailable'); }
  }
  async recover(): Promise<void> {
    if (this.disposed || !this.state.uncertain || this.state.busy) return;
    const generation = this.generation;
    this.state = { ...this.state, busy: true }; this.changed();
    try {
      const refreshed = await this.refresh();
      if (generation !== this.generation || this.disposed) return;
      this.state = { ...this.state, busy: false, uncertain: refreshed === false, error: refreshed === false ? 'Inventory could not be refreshed. Please try again.' : '' };
    } catch { if (generation === this.generation && !this.disposed) this.state = { ...this.state, busy: false, error: 'Inventory could not be refreshed. Please try again.' }; }
    if (!this.disposed) this.changed();
  }
  private refreshAfterChange(): void { void Promise.resolve().then(() => this.refresh()).catch(() => { if (!this.disposed) { this.state = { ...this.state, error: 'Your inventory could not be refreshed. Reload to see the latest changes.' }; this.changed(); } }); }
  stop(): void {
    if (this.disposed) return;
    if (this.approving) { this.fail('unavailable'); return; }
    this.retire(); this.state = { ...this.state, busy: false, plan: null }; this.changed();
  }
  dispose(): void {
    this.retire(); this.disposed = true;
    this.state = { messages: [], busy: false, plan: null, error: '', uncertain: false };
  }
  private retire(): void { this.generation++; this.controller?.abort(); this.connection?.close(); this.connection = undefined; }
  private fail(kind: 'authentication' | 'unavailable'): void {
    const uncertain = this.approving; this.approving = false; this.retire();
    this.state = { ...this.state, busy: false, plan: null, uncertain,
      error: uncertain ? 'The connection ended before the result was confirmed. Refresh your inventory before trying another change.' : kind === 'authentication' ? 'Sign in again to continue.' : 'The conversation could not continue. Please try again.' };
    if (uncertain) this.refreshAfterChange();
    this.changed(); if (kind === 'authentication') this.authenticationLost();
  }
  private receive(event: InventoryConversationEvent): void {
    if (event.type === 'answer') this.state = { ...this.state, messages: [...this.state.messages, { role: 'assistant', text: event.text, assets: event.assets }] };
    if (event.type === 'review') this.state = { ...this.state, busy: false, plan: event.plan };
    if (event.type === 'changed') { this.approving = false; this.state = { ...this.state, busy: false, plan: null }; this.refreshAfterChange(); }
    if (event.type === 'cancelled') { this.approving = false; this.state = { ...this.state, busy: false, plan: null }; }
    if (event.type === 'ready') this.state = { ...this.state, busy: false };
    if (event.type === 'ended') { this.retire(); this.state = { ...this.state, busy: false }; }
    this.changed();
  }
}
