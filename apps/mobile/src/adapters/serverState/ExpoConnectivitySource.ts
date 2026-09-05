import type { ConnectivitySource } from '../../application/shared/ConnectivitySource';

type NativeNetworkState = { readonly isConnected?: boolean };
export interface NativeConnectivityRuntime {
  getNetworkStateAsync(): Promise<NativeNetworkState>;
  addNetworkStateListener(listener: (state: NativeNetworkState) => void): { remove(): void };
}

export class ExpoConnectivitySource implements ConnectivitySource {
  constructor(private readonly runtime: NativeConnectivityRuntime) {}
  subscribe(onConnectivity: (connected: boolean) => void): () => void {
    let active = true; let receivedEvent = false;
    const subscription = this.runtime.addNetworkStateListener(state => {
      receivedEvent = true;
      if (active) onConnectivity(state.isConnected !== false);
    });
    void this.runtime.getNetworkStateAsync().then(state => {
      if (active && !receivedEvent) onConnectivity(state.isConnected !== false);
    }).catch(() => { /* Unknown connectivity must not block a reachable local server. */ });
    return () => { active = false; subscription.remove(); };
  }
}
