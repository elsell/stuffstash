export type NotificationSettingsPage =
  | { readonly kind: 'overview' }
  | { readonly kind: 'type'; readonly typeId: string }
  | { readonly kind: 'timing'; readonly typeId?: string }
  | { readonly kind: 'timezone' };
export type NotificationSettingsDestination = Exclude<NotificationSettingsPage, {kind:'overview'}>;
export function notificationSettingsEditorTarget(params: Record<string, unknown>): {
  page: NotificationSettingsDestination; scope: {tenantId:string;inventoryId:string};
} | null {
  const {view,typeId,tenantId,inventoryId}=params;
  if(typeof tenantId!=='string' || !tenantId || typeof inventoryId!=='string' || !inventoryId || (typeId!==undefined && (typeof typeId!=='string' || !typeId)))return null;
  const scope={tenantId,inventoryId};
  if(view==='timezone')return {page:{kind:'timezone'},scope};
  if(view==='timing')return {page:{kind:'timing',typeId},scope};
  if(view==='type' && typeId)return {page:{kind:'type',typeId},scope};
  return null;
}
