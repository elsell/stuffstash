import type { VoiceInventoryContextRepository } from './VoiceInventoryContext';
import type { ReadRequest } from '../shared/ReadRequest';

export type VoiceActionPreviewViewModel = {
  readonly summary: string;
  readonly steps: readonly string[];
  readonly riskLabel: string;
};

export type VoiceInteractionPreviewViewModel = {
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly sampleUtterance: string;
  readonly assistantSummary: string;
  readonly actionPreview: VoiceActionPreviewViewModel;
};

export class VoiceInteractionPreviewQuery {
  constructor(private readonly inventories: VoiceInventoryContextRepository) {}

  async execute(request: ReadRequest = {}): Promise<VoiceInteractionPreviewViewModel> {
    const context = await this.inventories.getVoiceInventoryContext(request);

    return {
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      sampleUtterance: 'Move the fertilizer from the garage shelf to the wire rack.',
      assistantSummary: 'I found one likely move. Review the plan before anything changes.',
      actionPreview: {
        summary: 'Move fertilizer',
        steps: [
          'Find Fertilizer in Garage shelf',
          'Move it to Wire rack in Garage',
          'Record the change in inventory history'
        ],
        riskLabel: 'Needs approval before saving'
      }
    };
  }
}
