// In-memory outgoing share surface for composition tests; native acceptance uses Expo.
const files: string[] = [];
export async function isAvailableAsync() { return true; }
export async function shareAsync(uri: string) { files.push(uri); }
export function sharedFiles() { return [...files]; }
export function resetSharedFiles() { files.length = 0; }
