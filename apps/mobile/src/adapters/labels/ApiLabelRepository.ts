import { LabelsClient, type LabelMedia as WireMedia } from '@stuff-stash/api-client';
import type { LabelFile, LabelMedia, LabelReference, LabelRepository, LabelScope, LabelSelection } from '../../application/labels/LabelWorkspace';
import { assertReadActive } from '../../application/shared/ReadRequest';

export class ApiLabelRepository implements LabelRepository {
  constructor(private readonly client: LabelsClient) {}
  async instance(signal: AbortSignal) { return (await this.client.instance(signal)).instanceId; }
  async resolve(reference: LabelReference, signal: AbortSignal) {
    const result = await this.client.resolve(reference.instanceId, reference.labelId, signal);
    return { tenantId: result.tenantId, inventoryId: result.inventoryId, assetId: result.assetId, archived: result.lifecycleState === 'archived' };
  }
  async catalog(scope: LabelScope, signal: AbortSignal) {
    const [templates, profiles] = await Promise.all([
      this.client.templates(scope.tenantId, scope.inventoryId, signal),
      this.client.profiles(scope.tenantId, scope.inventoryId, signal)
    ]);
    assertReadActive(signal);
    return {
      templates: (templates ?? []).map(template => ({ id: template.id, version: template.version, name: template.name, showReference: template.defaults.show_reference, supportsReference: (template.options ?? []).includes('show_reference') })),
      profiles: (profiles ?? []).flatMap(profile => (profile.media ?? []).map(media => ({ id: `${profile.adapterId}:${media.presetId}:${media.version}`, name: media.name, media: { presetId: media.presetId, version: media.version, widthMicrometers: media.widthMicrometers, heightMicrometers: media.heightMicrometers, margins: media.marginsMicrometers, resolutionDPI: media.resolutionDpi, rasterWidth: media.rasterWidth, rasterHeight: media.rasterHeight, orientation: media.orientation, colorMode: media.colorMode, cutPolicy: media.cutPolicy, displayRotation: media.displayRotation } })))
    };
  }
  async render(scope: LabelScope, assetId: string, selection: LabelSelection, format: 'png' | 'pdf', signal: AbortSignal): Promise<LabelFile> {
    await this.client.provision(scope.tenantId, scope.inventoryId, assetId, signal);
    assertReadActive(signal);
    const rendered = await this.client.render(scope.tenantId, scope.inventoryId, assetId, {
      template: { id: selection.template.id, version: selection.template.version, options: { show_reference: selection.showReference } },
      media: toWireMedia(selection.media), format
    }, signal);
    assertReadActive(signal);
    const bytes = await this.client.contentBytes(scope.tenantId, scope.inventoryId, rendered.id, format, signal);
    assertReadActive(signal);
    return { bytes, format, width: rendered.widthPixels, height: rendered.heightPixels, rotation: rendered.displayRotation };
  }
}
function toWireMedia(media: LabelMedia): WireMedia {
  return { preset_id: media.presetId, version: media.version, width_micrometers: media.widthMicrometers, height_micrometers: media.heightMicrometers,
    margins_micrometers: media.margins, resolution_dpi: media.resolutionDPI, raster_width: media.rasterWidth, raster_height: media.rasterHeight,
    orientation: media.orientation, color_mode: media.colorMode, cut_policy: media.cutPolicy, display_rotation: media.displayRotation };
}
