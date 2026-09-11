export const AndroidImportance = { DEFAULT: 3 };
export async function setNotificationChannelAsync() { return null; }
export async function getPermissionsAsync() { return { granted: false }; }
export async function requestPermissionsAsync() { return { granted: false }; }
export async function getDevicePushTokenAsync(): Promise<{type:string;data:unknown}> { throw new Error('Native push token is unavailable in this test environment.'); }
