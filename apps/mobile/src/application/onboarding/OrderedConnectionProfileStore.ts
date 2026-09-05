import type { ConnectionProfileStore, SavedConnectionProfile } from './ConnectionProfile';

// Reset must finish after any profile save already in progress, not race it.
export class OrderedConnectionProfileStore implements ConnectionProfileStore {
  private writes: Promise<void> = Promise.resolve();
  constructor(private readonly store: ConnectionProfileStore) {}
  load() { return this.store.load(); }
  save(profile: SavedConnectionProfile) { return this.write(() => this.store.save(profile)); }
  clear() { return this.write(() => this.store.clear()); }
  private write(action: () => Promise<void>) {
    const result = this.writes.then(action);
    this.writes = result.catch(() => {});
    return result;
  }
}
