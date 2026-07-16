export const appearancePreferences = ['system', 'light', 'dark'] as const;

export type AppearancePreference = typeof appearancePreferences[number];

export interface AppearancePreferenceStore {
  load(): Promise<AppearancePreference | undefined>;
  save(preference: AppearancePreference): Promise<void>;
}

export class AppearancePreferenceController {
  private pendingSave: Promise<void> = Promise.resolve();

  constructor(private readonly preferences: AppearancePreferenceStore) {}

  async getPreference(): Promise<AppearancePreference> {
    return (await this.preferences.load()) ?? 'system';
  }

  async setPreference(preference: AppearancePreference): Promise<void> {
    const save = this.pendingSave
      .catch(() => undefined)
      .then(() => this.preferences.save(preference));
    this.pendingSave = save;
    await save;
  }
}

export function isAppearancePreference(value: unknown): value is AppearancePreference {
  return typeof value === 'string' && appearancePreferences.includes(value as AppearancePreference);
}
