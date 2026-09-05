export const getNetworkStateAsync = async () => ({ isConnected: true });
export const addNetworkStateListener = (_listener: (state: { isConnected: boolean }) => void) => ({ remove() {} });
