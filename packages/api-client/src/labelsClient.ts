import {createAuthenticatedTransport} from './authenticatedTransport';
import {StuffStashAPIError, type StuffStashClientOptions} from './stuffStashClient';
import type {components} from './generated/schema';
export type LabelMedia = components['schemas']['Media'];
export type LabelTemplateSelection = components['schemas']['TemplateSelection'];
export type LabelRenderRequest = {media:LabelMedia; template:LabelTemplateSelection; format:'png'|'pdf'};
const scope = '/tenants/{tenantId}/inventories/{inventoryId}';
/** Transport infrastructure: all URLs derive from the configured API and generated paths. */
export class LabelsClient {
  private readonly transport;
  constructor(options:StuffStashClientOptions) { this.transport = createAuthenticatedTransport(options); }
  async instance(signal?:AbortSignal) { return unwrap(await this.transport.GET('/instance', {signal})).data; }
  async resolve(instanceId:string,labelId:string,signal?:AbortSignal) {
    return unwrap(await this.transport.GET('/labels/v1/{instanceId}/{labelId}', {params:{path:{instanceId,labelId}},signal})).data;
  }
  async provision(tenantId:string,inventoryId:string,assetId:string,signal?:AbortSignal) {
    return unwrap(await this.transport.POST(`${scope}/assets/{assetId}/label`, {params:{path:{tenantId,inventoryId,assetId}},signal})).data;
  }
  async templates(tenantId:string,inventoryId:string,signal?:AbortSignal) {
    return unwrap(await this.transport.GET(`${scope}/label-templates`, {params:{path:{tenantId,inventoryId}},signal})).data;
  }
  async profiles(tenantId:string,inventoryId:string,signal?:AbortSignal) {
    return unwrap(await this.transport.GET(`${scope}/printer-profiles`, {params:{path:{tenantId,inventoryId}},signal})).data;
  }
  async render(tenantId:string,inventoryId:string,assetId:string,body:LabelRenderRequest,signal?:AbortSignal) {
    return unwrap(await this.transport.POST(`${scope}/assets/{assetId}/label-renders`, {params:{path:{tenantId,inventoryId,assetId}},body,signal})).data;
  }
  /** Native runtimes cannot construct a Blob from Expo's ArrayBuffer response. */
  async contentBytes(tenantId:string,inventoryId:string,renderId:string,format:'png'|'pdf',signal?:AbortSignal):Promise<Uint8Array> {
    const result = await this.transport.GET(`${scope}/label-renders/{renderId}/content`, {params:{path:{tenantId,inventoryId,renderId}},parseAs:'arrayBuffer',signal});
    const content = unwrap(result);
    const expected = format === 'png' ? 'image/png' : 'application/pdf';
    if (result.response.headers.get('Content-Type')?.split(';')[0].trim().toLowerCase() !== expected) {
      throw new StuffStashAPIError(502,'invalid_label','Invalid label content.');
    }
    return new Uint8Array(content);
  }
  async content(tenantId:string,inventoryId:string,renderId:string,signal?:AbortSignal):Promise<Blob> {
    return unwrap(await this.transport.GET(`${scope}/label-renders/{renderId}/content`, {params:{path:{tenantId,inventoryId,renderId}},parseAs:'blob',signal}));
  }
}
function unwrap<T>(result:{data?:T;error?:unknown;response:Response}):T {
  if (!result.response.ok || result.error || result.data === undefined) {
    throw new StuffStashAPIError(result.response.status,'label_unavailable','Label unavailable.');
  }
  return result.data;
}
