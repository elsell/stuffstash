export type LabelReference = { readonly instanceId: string; readonly labelId: string };
export type LabelScope = { readonly tenantId: string; readonly inventoryId: string };
export type LabelAsset = LabelScope & { readonly assetId: string; readonly archived: boolean };
export type LabelMedia = {
  readonly presetId?: string; readonly version?: number;
  readonly widthMicrometers: number; readonly heightMicrometers: number;
  readonly margins: { readonly left: number; readonly right: number; readonly top: number; readonly bottom: number };
  readonly resolutionDPI: number; readonly rasterWidth: number; readonly rasterHeight: number;
  readonly orientation: string; readonly colorMode: string; readonly cutPolicy: string; readonly displayRotation: number;
};
export type LabelProfile = { readonly id: string; readonly name: string; readonly media: LabelMedia };
export type LabelTemplate = { readonly id: string; readonly version: number; readonly name: string; readonly showReference: boolean; readonly supportsReference: boolean };
export type LabelSelection = { readonly template: LabelTemplate; readonly media: LabelMedia; readonly showReference: boolean };
export type LabelFile = { readonly bytes: Uint8Array; readonly format: 'png' | 'pdf'; readonly width: number; readonly height: number; readonly rotation: number };
export interface LabelRepository {
  instance(signal: AbortSignal): Promise<string>;
  resolve(reference: LabelReference, signal: AbortSignal): Promise<LabelAsset>;
  catalog(scope: LabelScope, signal: AbortSignal): Promise<{ profiles: readonly LabelProfile[]; templates: readonly LabelTemplate[] }>;
  render(scope: LabelScope, assetId: string, selection: LabelSelection, format: 'png' | 'pdf', signal: AbortSignal): Promise<LabelFile>;
}
export interface LabelFiles {
  preview(file: LabelFile, signal: AbortSignal): Promise<{ uri: string; release(): void }>;
  deliver(file: LabelFile, action: 'share' | 'print', signal: AbortSignal): Promise<void>;
}
export type LabelWorkspace = { readonly repository: LabelRepository; readonly files: LabelFiles; readonly parse: (source: string) => LabelReference };
export class LabelFailure extends Error {
  constructor(readonly code: 'wrong_instance' | 'invalid_label') { super(code); }
}
