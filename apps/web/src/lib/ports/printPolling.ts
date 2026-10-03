export interface PrintPollingRuntime {
    visible(): boolean;
    subscribe(listener: () => void): () => void;
    schedule(delayMs: number, callback: () => void): () => void;
}
