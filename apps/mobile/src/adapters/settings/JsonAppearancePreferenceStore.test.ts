import { describe, expect, it } from 'vitest';
import type { AppearancePreferenceTextFile } from './JsonAppearancePreferenceStore';
import { JsonAppearancePreferenceStore } from './JsonAppearancePreferenceStore';

class FakeAppearancePreferenceTextFile implements AppearancePreferenceTextFile {
  content: string | undefined;

  async read(): Promise<string | undefined> {
    return this.content;
  }

  async write(content: string): Promise<void> {
    this.content = content;
  }
}

describe('JsonAppearancePreferenceStore', () => {
  it('round trips a typed appearance preference', async () => {
    const file = new FakeAppearancePreferenceTextFile();
    const store = new JsonAppearancePreferenceStore(file);

    await store.save('dark');

    expect(file.content).toBe(JSON.stringify({ preference: 'dark' }));
    await expect(store.load()).resolves.toBe('dark');
  });

  it('ignores missing, malformed, and unsupported preferences', async () => {
    const file = new FakeAppearancePreferenceTextFile();
    const store = new JsonAppearancePreferenceStore(file);

    await expect(store.load()).resolves.toBeUndefined();
    file.content = '{not json';
    await expect(store.load()).resolves.toBeUndefined();
    file.content = JSON.stringify({ preference: 'sepia' });
    await expect(store.load()).resolves.toBeUndefined();
  });
});
