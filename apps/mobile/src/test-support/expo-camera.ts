import { useState } from 'react';
export const CameraView = 'CameraView';
type Permission = { granted: boolean; canAskAgain: boolean };
class CameraPermissionsFake {
  allowed = true;
  requests = 0;
  private deferred = false;
  private pending: Array<(permission: Permission) => void> = [];
  deferDecision() { this.deferred = true; }
  decide(allowed: boolean) {
    this.allowed = allowed;
    this.deferred = false;
    for (const resolve of this.pending.splice(0)) resolve(this.current());
  }
  current(): Permission { return { granted: this.allowed, canAskAgain: true }; }
  request(): Promise<Permission> {
    this.requests++;
    return this.deferred ? new Promise(resolve => this.pending.push(resolve)) : Promise.resolve(this.current());
  }
  reset() { this.decide(true); this.requests = 0; }
}
export const cameraPermissionsFake = new CameraPermissionsFake();
export function setCameraAllowed(value: boolean) { cameraPermissionsFake.allowed = value; }
export function useCameraPermissions() {
  const [permission, setPermission] = useState(cameraPermissionsFake.current());
  return [permission, async () => { const next = await cameraPermissionsFake.request(); setPermission(next); return next; }] as const;
}
