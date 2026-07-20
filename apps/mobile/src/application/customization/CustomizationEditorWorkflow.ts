export class CustomizationEditorWorkflow {
  private loadGeneration = 0;
  private savePhase: 'idle' | 'saving' = 'idle';
  private lifecyclePhase: 'idle' | 'confirming' | 'mutating' = 'idle';
  private exitPhase: 'blocked' | 'authorized' | 'dispatched' = 'blocked';
  private exitAction: unknown;

  beginLoad(): number { this.loadGeneration += 1; return this.loadGeneration; }
  invalidateLoads(): void { this.loadGeneration += 1; }
  isCurrentLoad(generation: number): boolean { return generation === this.loadGeneration; }

  beginSave(): boolean {
    if (this.savePhase !== 'idle' || this.lifecyclePhase !== 'idle') return false;
    this.savePhase = 'saving';
    return true;
  }
  finishSave(): void { this.savePhase = 'idle'; }

  beginLifecycleConfirmation(): boolean {
    if (this.lifecyclePhase !== 'idle' || this.savePhase !== 'idle') return false;
    this.lifecyclePhase = 'confirming';
    return true;
  }
  beginLifecycleMutation(): boolean {
    if (this.lifecyclePhase !== 'confirming') return false;
    this.lifecyclePhase = 'mutating';
    return true;
  }
  cancelLifecycleConfirmation(): boolean {
    if (this.lifecyclePhase !== 'confirming') return false;
    this.lifecyclePhase = 'idle';
    return true;
  }
  finishLifecycle(): void { this.lifecyclePhase = 'idle'; }

  authorizeExit(action: unknown): boolean {
    if (this.exitPhase !== 'blocked') return false;
    this.exitAction = action;
    this.exitPhase = 'authorized';
    return true;
  }
  takeAuthorizedExit(): unknown {
    if (this.exitPhase !== 'authorized') return undefined;
    const action = this.exitAction;
    this.exitAction = undefined;
    this.exitPhase = 'dispatched';
    return action;
  }
  resetExit(): void { this.exitAction = undefined; this.exitPhase = 'blocked'; }
}
