import { describe, expect, it } from 'vitest';
import {
  ConnectionProfileTextFile,
  JsonConnectionProfileStore
} from './JsonConnectionProfileStore';

class FakeConnectionProfileTextFile implements ConnectionProfileTextFile {
  content: string | undefined;

  async read(): Promise<string | undefined> {
    return this.content;
  }

  async write(content: string): Promise<void> {
    this.content = content;
  }

  async delete(): Promise<void> {
    this.content = undefined;
  }
}

describe('JsonConnectionProfileStore', () => {
  it('returns undefined when no saved profile exists', async () => {
    const file = new FakeConnectionProfileTextFile();
    const store = new JsonConnectionProfileStore(file);

    await expect(store.load()).resolves.toBeUndefined();
  });

  it('persists only non-secret connection profile metadata', async () => {
    const file = new FakeConnectionProfileTextFile();
    const store = new JsonConnectionProfileStore(file);

    await store.save({
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-home'
    });

    expect(file.content).toBe(
      JSON.stringify({
        apiBaseUrl: 'http://localhost:8080',
        tenantId: 'tenant-home'
      })
    );
    expect(file.content).not.toContain('devToken');
    await expect(store.load()).resolves.toEqual({
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-home'
    });
  });

  it('whitelists persisted fields even when passed a richer object', async () => {
    const file = new FakeConnectionProfileTextFile();
    const store = new JsonConnectionProfileStore(file);

    await store.save({
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-home',
      devToken: 'dev:user-1'
    } as Parameters<JsonConnectionProfileStore['save']>[0]);

    expect(file.content).toBe(
      JSON.stringify({
        apiBaseUrl: 'http://localhost:8080',
        tenantId: 'tenant-home'
      })
    );
  });

  it('normalizes saved URLs and tenant IDs on load', async () => {
    const file = new FakeConnectionProfileTextFile();
    file.content = JSON.stringify({
      apiBaseUrl: ' stuffstash.example.test/ ',
      tenantId: ' tenant-home '
    });
    const store = new JsonConnectionProfileStore(file);

    await expect(store.load()).resolves.toEqual({
      apiBaseUrl: 'https://stuffstash.example.test',
      tenantId: 'tenant-home'
    });
  });

  it('ignores malformed or incomplete saved profile JSON', async () => {
    const file = new FakeConnectionProfileTextFile();
    const store = new JsonConnectionProfileStore(file);

    file.content = '{not json';
    await expect(store.load()).resolves.toBeUndefined();

    file.content = JSON.stringify({ tenantId: 'tenant-home' });
    await expect(store.load()).resolves.toBeUndefined();
  });

  it('clears the saved profile idempotently', async () => {
    const file = new FakeConnectionProfileTextFile();
    file.content = JSON.stringify({ apiBaseUrl: 'http://localhost:8080' });
    const store = new JsonConnectionProfileStore(file);

    await store.clear();
    await store.clear();

    expect(file.content).toBeUndefined();
  });
});
