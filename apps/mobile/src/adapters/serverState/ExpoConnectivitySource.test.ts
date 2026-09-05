import { expect, it } from 'vitest';
import { ExpoConnectivitySource } from './ExpoConnectivitySource';
it('uses native connection events, ignores obsolete startup state and cleans up', async () => {
  let initial!: (state: { isConnected: boolean }) => void;
  let listener!: (state: { isConnected?: boolean; isInternetReachable?: boolean }) => void; let removed = false;
  const source = new ExpoConnectivitySource({ getNetworkStateAsync: () => new Promise(resolve => { initial = resolve; }), addNetworkStateListener: callback => { listener = callback; return { remove: () => { removed = true; } }; } });
  const values: boolean[] = []; const stop = source.subscribe(value => values.push(value));
  listener({ isConnected: false }); initial({ isConnected: true }); await Promise.resolve();
  expect(values).toEqual([false]);
  listener({ isConnected: true, isInternetReachable: false }); // Local servers can work without public Internet.
  expect(values).toEqual([false, true]); stop(); expect(removed).toBe(true);
  listener({ isConnected: false }); expect(values).toEqual([false, true]);
});
