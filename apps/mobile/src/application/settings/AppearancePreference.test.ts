import { describe, expect, it, vi } from 'vitest';
import {
  AppearancePreferenceController,
  appearancePreferences,
  type AppearancePreference,
  type AppearancePreferenceStore
} from './AppearancePreference';

class FakeAppearancePreferenceStore implements AppearancePreferenceStore {
  saved: AppearancePreference | undefined;

  async load(): Promise<AppearancePreference | undefined> {
    return this.saved;
  }

  async save(preference: AppearancePreference): Promise<void> {
    this.saved = preference;
  }
}

describe('AppearancePreferenceController', () => {
  it('presents the system-following option first', () => {
    expect(appearancePreferences).toEqual(['system', 'light', 'dark']);
  });

  it('defaults to the system appearance when no preference has been saved', async () => {
    const controller = new AppearancePreferenceController(new FakeAppearancePreferenceStore());

    await expect(controller.getPreference()).resolves.toBe('system');
  });

  it('loads and changes the device-local preference through its port', async () => {
    const store = new FakeAppearancePreferenceStore();
    store.saved = 'dark';
    const controller = new AppearancePreferenceController(store);

    await expect(controller.getPreference()).resolves.toBe('dark');
    await controller.setPreference('light');

    expect(store.saved).toBe('light');
  });

  it('serializes rapid preference saves so the last selection persists last', async () => {
    const writes: AppearancePreference[] = [];
    const releases: Array<() => void> = [];
    const store: AppearancePreferenceStore = {
      async load() { return undefined; },
      async save(preference) {
        writes.push(preference);
        await new Promise<void>((resolve) => releases.push(resolve));
      }
    };
    const controller = new AppearancePreferenceController(store);

    const darkSave = controller.setPreference('dark');
    const lightSave = controller.setPreference('light');
    await vi.waitFor(() => expect(writes).toEqual(['dark']));

    releases.shift()?.();
    await darkSave;
    await vi.waitFor(() => expect(writes).toEqual(['dark', 'light']));

    releases.shift()?.();
    await lightSave;
    expect(writes.at(-1)).toBe('light');
  });
});
