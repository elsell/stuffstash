import { expect, it } from 'vitest';
import { NotificationPreferencesSession } from './NotificationPreferencesSession';
import type { NotificationPreferences, NotificationPreferencesUpdate, ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import type { NotificationEvent } from './NotificationObservability';
class InMemoryWorkspaceObserver { events: NotificationEvent[] = []; record(event: NotificationEvent) { this.events.push(event); } }

class PreferencesFake {
  value: NotificationPreferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'America/New_York', pushEnabled: true, overrides: [] };
  writes: NotificationPreferencesUpdate[] = [];
  async initializePreferences() { return structuredClone(this.value); }
  async getPreferences() { return structuredClone(this.value); }
  async updatePreferences(_tenant: string, _inventory: string, input: NotificationPreferencesUpdate) {
    if (input.revision !== this.value.revision) throw new Error('Conflict');
    this.writes.push(structuredClone(input));
    this.value = { ...input, revision: input.revision + 1, overrides: this.value.overrides };
    return structuredClone(this.value);
  }
  async setTypeOverride(_tenant: string, _inventory: string, typeId: string, revision: number, settings: ExpirationReminderPolicy) {
    if (revision !== this.value.revision) throw new Error('Conflict');
    this.value.overrides = [...this.value.overrides.filter((value) => value.customAssetTypeId !== typeId), { customAssetTypeId: typeId, settings }];
    this.value.revision++;
    return structuredClone(this.value);
  }
  async removeTypeOverride(_tenant: string, _inventory: string, typeId: string, revision: number) {
    if (revision !== this.value.revision) throw new Error('Conflict');
    this.value.overrides = this.value.overrides.filter((value) => value.customAssetTypeId !== typeId);
    this.value.revision++;
    return structuredClone(this.value);
  }
}
it('preserves delivery settings across policy edits and advances revisions across overrides', async () => {
  const repository = new PreferencesFake();
  const session = new NotificationPreferencesSession(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory');
  await session.initialize('UTC');
  await session.saveDefaults({ ...repository.value.defaults, enabled: false });
  expect(repository.writes[0]).toMatchObject({ revision: 1, timezone: 'America/New_York', pushEnabled: true, defaults: { enabled: false } });
  await session.saveTypeOverride('medicine', { ...repository.value.defaults, enabled: true });
  expect(session.snapshot?.overrides[0].settings.enabled).toBe(true);
  await session.saveTypeOverride('medicine', null);
  expect(session.snapshot).toMatchObject({ revision: 4, overrides: [] });
});
it('keeps its snapshot after conflict and requires refresh before retrying', async () => {
  const repository = new PreferencesFake();
  const observer = new InMemoryWorkspaceObserver();
  const session = new NotificationPreferencesSession(repository, observer, 'tenant', 'inventory');
  await session.initialize('UTC');
  repository.value.revision++;
  repository.value.pushEnabled = false;
  await expect(session.saveDefaults({ ...repository.value.defaults, advanceDays: 10 })).rejects.toThrow('Conflict');
  expect(session.snapshot).toMatchObject({ revision: 1, pushEnabled: true });
  await session.refresh();
  await session.saveDefaults({ ...repository.value.defaults, advanceDays: 10 });
  expect(session.snapshot).toMatchObject({ revision: 3, pushEnabled: false, defaults: { advanceDays: 10 } });
  expect(observer.events.some((event) => event.outcome === 'failed')).toBe(true);
});
it('rejects overlapping saves and preserves overrides during timezone changes', async () => {
  const repository = new PreferencesFake();
  const session = new NotificationPreferencesSession(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory');
  await session.initialize('UTC');
  await session.saveTypeOverride('food', { ...repository.value.defaults, advanceDays: 3 });
  const first = session.saveTimezone('Europe/London');
  await expect(session.saveDefaults(repository.value.defaults)).rejects.toThrow('already in progress');
  await first;
  expect(repository.writes).toHaveLength(1);
  expect(session.snapshot).toMatchObject({ timezone: 'Europe/London', pushEnabled: true, overrides: [{ customAssetTypeId: 'food', settings: { advanceDays: 3 } }] });
  const snapshot = session.snapshot!;
  snapshot.defaults.enabled = false;
  expect(session.snapshot?.defaults.enabled).toBe(true);
});
it('changes push preference without changing reminder policy and preserves snapshots on cancellation', async () => {
  const repository = new PreferencesFake();
  const session = new NotificationPreferencesSession(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory');
  await session.initialize('UTC');
  await session.savePushEnabled(false);
  expect(session.snapshot).toMatchObject({ pushEnabled: false, defaults: { advanceDays: 30 }, timezone: 'America/New_York' });
  const controller = new AbortController();
  const cancelling = new NotificationPreferencesSession({ ...repository,
    initializePreferences: repository.initializePreferences.bind(repository),
    updatePreferences: repository.updatePreferences.bind(repository),
    setTypeOverride: repository.setTypeOverride.bind(repository),
    removeTypeOverride: repository.removeTypeOverride.bind(repository),
    async getPreferences() { controller.abort(); return repository.value; }
  }, new InMemoryWorkspaceObserver(), 'tenant', 'inventory');
  await cancelling.initialize('UTC');
  const before = cancelling.snapshot;
  await expect(cancelling.refresh({ signal: controller.signal })).rejects.toMatchObject({ name: 'AbortError' });
  expect(cancelling.snapshot).toEqual(before);
});
