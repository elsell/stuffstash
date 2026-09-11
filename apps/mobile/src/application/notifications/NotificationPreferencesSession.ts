import type { NotificationPreferences, ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import type { NotificationRepository } from './NotificationRepository';
import type { NotificationObservability } from './NotificationObservability';
import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';

type PreferencesRepository = Pick<NotificationRepository, 'getPreferences' | 'initializePreferences' | 'updatePreferences' | 'setTypeOverride' | 'removeTypeOverride'>;

export class NotificationPreferencesSession {
  private current: NotificationPreferences | null = null;
  private pending = false;
  constructor(private readonly repository: PreferencesRepository, private readonly observer: NotificationObservability, private readonly tenantId: string, private readonly inventoryId: string) {}
  get snapshot(): NotificationPreferences | null { return this.current ? copyPreferences(this.current) : null; }
  initialize(timezone: string, request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => this.repository.initializePreferences(this.tenantId, this.inventoryId, timezone, request.signal));
  }
  refresh(request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => this.repository.getPreferences(this.tenantId, this.inventoryId, request.signal));
  }
  saveDefaults(defaults: ExpirationReminderPolicy, request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => {
      const current = this.loaded();
      return this.repository.updatePreferences(this.tenantId, this.inventoryId, {
        revision: current.revision, defaults, timezone: current.timezone, pushEnabled: current.pushEnabled
      }, request.signal);
    });
  }
  saveTimezone(timezone: string, request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => {
      const current = this.loaded();
      return this.repository.updatePreferences(this.tenantId, this.inventoryId, {
        revision: current.revision, defaults: current.defaults, timezone, pushEnabled: current.pushEnabled
      }, request.signal);
    });
  }
  savePushEnabled(pushEnabled: boolean, request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => {
      const current = this.loaded();
      return this.repository.updatePreferences(this.tenantId, this.inventoryId, {
        revision: current.revision, defaults: current.defaults, timezone: current.timezone, pushEnabled
      }, request.signal);
    });
  }
  saveTypeOverride(typeId: string, policy: ExpirationReminderPolicy | null, request: ReadRequest = {}): Promise<NotificationPreferences> {
    return this.run(request, () => {
      const revision = this.loaded().revision;
      return policy === null
        ? this.repository.removeTypeOverride(this.tenantId, this.inventoryId, typeId, revision, request.signal)
        : this.repository.setTypeOverride(this.tenantId, this.inventoryId, typeId, revision, policy, request.signal);
    });
  }
  private loaded(): NotificationPreferences {
    if (!this.current) throw new Error('Load reminder settings before saving.');
    return this.current;
  }
  private async run(request: ReadRequest, operation: () => Promise<NotificationPreferences>): Promise<NotificationPreferences> {
    assertReadActive(request.signal);
    if (this.pending) throw new Error('A reminder settings request is already in progress.');
    this.pending = true;
    try {
      const result = await operation();
      assertReadActive(request.signal);
      this.current = copyPreferences(result);
      this.observer.record({ operation: 'preferences', outcome: 'succeeded' });
      return copyPreferences(result);
    } catch (error) {
      this.observer.record({ operation: 'preferences', outcome: 'failed' });
      throw error;
    } finally { this.pending = false; }
  }
}
function copyPreferences(value: NotificationPreferences): NotificationPreferences {
  return { ...value, defaults: { ...value.defaults }, overrides: value.overrides.map((entry) => ({ ...entry, settings: { ...entry.settings } })) };
}
