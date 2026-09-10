import type { NotificationPreferences, ExpirationReminderPolicy } from '$lib/domain/notification';
import type { NotificationRepository } from '$lib/ports/notificationRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';

type PreferencesRepository = Pick<NotificationRepository, 'getPreferences' | 'initializePreferences' | 'updatePreferences' | 'setTypeOverride' | 'removeTypeOverride'>;

export class NotificationPreferencesSession {
  private current: NotificationPreferences | null = null;
  private pending = false;
  constructor(private readonly repository: PreferencesRepository, private readonly observer: WorkspaceObserver, private readonly tenantId: string, private readonly inventoryId: string) {}
  get snapshot(): NotificationPreferences | null { return this.current ? structuredClone(this.current) : null; }
  initialize(timezone: string): Promise<NotificationPreferences> {
    return this.run(false, () => this.repository.initializePreferences(this.tenantId, this.inventoryId, timezone));
  }
  refresh(): Promise<NotificationPreferences> {
    return this.run(false, () => this.repository.getPreferences(this.tenantId, this.inventoryId));
  }
  saveDefaults(defaults: ExpirationReminderPolicy): Promise<NotificationPreferences> {
    return this.run(true, () => {
      const current = this.loaded();
      return this.repository.updatePreferences(this.tenantId, this.inventoryId, {
        revision: current.revision, defaults, timezone: current.timezone, pushEnabled: current.pushEnabled
      });
    });
  }
  saveTimezone(timezone: string): Promise<NotificationPreferences> {
    return this.run(true, () => {
      const current = this.loaded();
      return this.repository.updatePreferences(this.tenantId, this.inventoryId, {
        revision: current.revision, defaults: current.defaults, timezone, pushEnabled: current.pushEnabled
      });
    });
  }
  saveTypeOverride(typeId: string, policy: ExpirationReminderPolicy | null): Promise<NotificationPreferences> {
    return this.run(true, () => {
      const revision = this.loaded().revision;
      return policy === null
        ? this.repository.removeTypeOverride(this.tenantId, this.inventoryId, typeId, revision)
        : this.repository.setTypeOverride(this.tenantId, this.inventoryId, typeId, revision, policy);
    });
  }
  private loaded(): NotificationPreferences {
    if (!this.current) throw new Error('Load reminder settings before saving.');
    return this.current;
  }
  private async run(mutation: boolean, operation: () => Promise<NotificationPreferences>): Promise<NotificationPreferences> {
    if (this.pending) throw new Error('A reminder settings request is already in progress.');
    this.pending = true;
    const attributes = { resource: 'notification_preferences', scope: 'inventory' };
    this.observer.record(mutation ? 'workspace.settings_mutation_started' : 'workspace.settings_collection_load_started', attributes);
    try {
      const result = await operation();
      this.current = structuredClone(result);
      this.observer.record(mutation ? 'workspace.settings_mutation_succeeded' : 'workspace.settings_collection_loaded', attributes);
      return structuredClone(result);
    } catch (error) {
      this.observer.record(mutation ? 'workspace.settings_mutation_failed' : 'workspace.settings_collection_load_failed', attributes);
      throw error;
    } finally { this.pending = false; }
  }
}
