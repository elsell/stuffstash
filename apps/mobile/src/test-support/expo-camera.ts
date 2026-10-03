import { useState } from 'react';
export const CameraView = 'CameraView';
let allowed = true;
export function setCameraAllowed(value: boolean) { allowed = value; }
export function useCameraPermissions() {
  const [permission, setPermission] = useState({ granted: allowed, canAskAgain: true });
  return [permission, async () => { const next = { granted: allowed, canAskAgain: true }; setPermission(next); return next; }] as const;
}
