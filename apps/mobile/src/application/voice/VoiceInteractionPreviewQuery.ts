import { t } from '../../presentation/localization';
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
      sampleUtterance: t('mobile.VoiceInteractionPreviewQuery.moveTheFertilizerFromTheGarageShelfToThe'),
      assistantSummary: t('mobile.VoiceInteractionPreviewQuery.iFoundOneLikelyMoveReviewThePlanBefore'),
      actionPreview: {
        summary: t('mobile.VoiceInteractionPreviewQuery.moveFertilizer'),
        steps: [
          t('mobile.VoiceInteractionPreviewQuery.findFertilizerInGarageShelf'),
          t('mobile.VoiceInteractionPreviewQuery.moveItToWireRackInGarage'),
          t('mobile.VoiceInteractionPreviewQuery.recordTheChangeInInventoryHistory')
        ],
        riskLabel: t('mobile.VoiceInteractionPreviewQuery.needsApprovalBeforeSaving')
      }
    };
  }
}
