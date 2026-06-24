export type WorkspaceEventName =
  | 'workspace.load_started'
  | 'workspace.load_failed'
  | 'workspace.loaded'
  | 'workspace.asset_create_started'
  | 'workspace.asset_created'
  | 'workspace.asset_create_failed'
  | 'workspace.asset_detail_load_started'
  | 'workspace.asset_detail_loaded'
  | 'workspace.asset_detail_load_failed'
  | 'workspace.asset_update_started'
  | 'workspace.asset_updated'
  | 'workspace.asset_update_failed'
  | 'workspace.search_started'
  | 'workspace.search_failed'
  | 'workspace.search_completed';

export interface WorkspaceObserver {
  record(eventName: WorkspaceEventName, attributes?: Record<string, string | number | boolean>): void;
}

export class InMemoryWorkspaceObserver implements WorkspaceObserver {
  readonly events: Array<{ eventName: WorkspaceEventName; attributes: Record<string, string | number | boolean> }> = [];

  record(eventName: WorkspaceEventName, attributes: Record<string, string | number | boolean> = {}): void {
    this.events.push({ eventName, attributes });
  }
}
