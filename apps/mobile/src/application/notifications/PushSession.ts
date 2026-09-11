import type { CurrentPrincipalRepository } from '../settings/SettingsQuery';
import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';
import { NotificationFailure } from './NotificationFailure';
import type { PushSetup } from './PushSetup';

type Preferences = Parameters<PushSetup['enable']>[1];
export class PushSession {
  private busy = false;
  private disconnected = false;
  private background: {controller: AbortController; work: Promise<void>} | null = null;
  constructor(
    private readonly serverId: string,
    private readonly principals: CurrentPrincipalRepository,
    private readonly setup: PushSetup,
    private readonly preferences: (tenantId: string, inventoryId: string) => Preferences
  ) {}

  async enable(tenantId: string, inventoryId: string, request: ReadRequest = {}) {
    return this.run(async () => {
      assertReadActive(request.signal);
      const principal = await this.principals.getCurrentPrincipal(request);
      assertReadActive(request.signal);
      return this.setup.enable({ serverId: this.serverId, principalId: principal.id, tenantId, inventoryId }, this.preferences(tenantId, inventoryId), request);
    });
  }

  async reconcile(request: ReadRequest = {}): Promise<void> {
    assertReadActive(request.signal);
    if(this.busy || this.disconnected) throw new NotificationFailure('conflict');
    const controller=new AbortController();
    const abort=()=>controller.abort();
    request.signal?.addEventListener('abort',abort,{once:true});
    const backgroundRequest={signal:controller.signal};
    const work=this.run(async()=>{
      if(!await this.setup.hasRegistrations(this.serverId,backgroundRequest)) return;
      const principal=await this.principals.getCurrentPrincipal(backgroundRequest);
      assertReadActive(controller.signal);
      await this.setup.reconcile(this.serverId,principal.id,backgroundRequest);
    });
    this.background={controller,work};
    try {await work;} finally {this.background=null;request.signal?.removeEventListener('abort',abort);}
  }

  async disconnect(action: () => Promise<void>, request: ReadRequest = {}): Promise<void> {
    if(this.background){
      this.background.controller.abort();
      try {await this.background.work;} catch { /* Cleanup below rechecks the journal. */ }
    }
    return this.run(async () => {
      assertReadActive(request.signal);
      if (await this.setup.hasRegistrations(this.serverId, request)) {
        const principal = await this.principals.getCurrentPrincipal(request);
        assertReadActive(request.signal);
        await this.setup.cleanup(this.serverId, principal.id, request);
      }
      assertReadActive(request.signal);
      await action();
      this.disconnected = true;
    });
  }

  private async run<T>(operation: () => Promise<T>): Promise<T> {
    if (this.busy || this.disconnected) throw new NotificationFailure('conflict');
    this.busy = true;
    try { return await operation(); } finally { this.busy = false; }
  }
}
