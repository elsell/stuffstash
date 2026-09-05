export interface ConnectivitySource {
  subscribe(onConnectivity: (connected: boolean) => void): () => void;
}
