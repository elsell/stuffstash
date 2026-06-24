import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

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
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(): Promise<VoiceInteractionPreviewViewModel> {
    const workspace = await this.inventories.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('Inventory workspace must include at least one inventory.');
    }

    const tenant = workspace.tenants.find((item) => item.id === inventory.tenantId);

    if (!tenant) {
      throw new Error('Selected inventory must belong to a tenant.');
    }

    return {
      tenantName: tenant.name,
      inventoryName: inventory.name,
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
