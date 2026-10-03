import { PrintRequestRejected } from '../printing/PrintSubmission';
import type { CreatePrintRequest } from '../printing/PrintingWorkspace';
import type { CreateInventoryAssetInput } from '../home/InventorySummaryRepository';
import { CatalogRecoveryError } from '../shared/CatalogRecoveryError';
import { t } from "../../presentation/localization";
import { assetId, type AssetKind, type AssetExpiration } from '../../domain/assets/AssetSummary';
import type { ActiveAssetTagReference, CreateAssetTagDraft } from '../assets/AssetTagDraftResolution';
import { createPendingAssetTags, reconcilePendingAssetTagDrafts } from '../assets/AssetTagDraftResolution';
import type {
  CreateInventoryAssetPhotoInput,
  InventorySummaryRepository
} from '../home/InventorySummaryRepository';

export type CreateAssetCommandInput = {
  readonly printRequest?: CreatePrintRequest;
  readonly expiration?: AssetExpiration;
  readonly customAssetTypeId?: string;
  readonly kind?: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
  readonly activeTags?: readonly ActiveAssetTagReference[];
  readonly photos?: readonly CreateAssetPhotoInput[];
};

export type CreateAssetPhotoInput = CreateInventoryAssetPhotoInput;

export type CreateAssetCommandResult = {
  readonly printJobId?: string;
  readonly id: string;
  readonly title: string;
  readonly message: string;
};

export class CreateAssetCommand {
  private readonly printAttempts = new Map<string, { input: CreateInventoryAssetInput; ambiguous: boolean }>();
  constructor(private readonly inventories: Pick<InventorySummaryRepository, 'createAsset' | 'createAssetTag' | 'addAssetPhoto'>) {}

  async execute(input: CreateAssetCommandInput): Promise<CreateAssetCommandResult> {
    const title = input.title.trim();
    if (title.length === 0) {
      throw new CatalogRecoveryError('recovery.nameRequired');
    }

    const requestIdentity = input.printRequest ? JSON.stringify([input.printRequest.scope, input.printRequest.key]) : undefined;
    let attempt = requestIdentity ? this.printAttempts.get(requestIdentity) : undefined;
    if (!attempt) {
      const reconciledTags = reconcilePendingAssetTagDrafts({ selectedTagIds: input.tagIds ?? [], pendingTags: input.newTags ?? [], activeTags: input.activeTags ?? [] });
      const createdTagIds = await createPendingAssetTags(this.inventories, reconciledTags.pendingTags, input.printRequest?.scope).catch(error => { if (input.printRequest) throw new PrintRequestRejected(); throw error; });
      const tagIds = [...reconciledTags.tagIds, ...createdTagIds];
      attempt = { ambiguous: false, input: {
        kind: input.kind ?? 'item', expiration: input.expiration, customAssetTypeId: input.customAssetTypeId,
        title, description: input.description.trim(), parentAssetId: input.parentAssetId ? assetId(input.parentAssetId) : undefined,
        tagIds: tagIds.length > 0 ? tagIds : undefined, printRequest: input.printRequest
      } };
      if (requestIdentity) this.printAttempts.set(requestIdentity, attempt);
    }
    let asset;
    try { asset = await this.inventories.createAsset(attempt.input); }
    catch (error) {
      if (requestIdentity) {
        if (error instanceof PrintRequestRejected && !attempt.ambiguous) this.printAttempts.delete(requestIdentity);
        else { attempt.ambiguous = true; if (error instanceof PrintRequestRejected) throw new Error('Earlier create outcome is unknown'); }
      }
      throw error;
    }
    let failedPhotoCount = 0;
    for (const photo of input.photos ?? []) {
      try {
        await this.inventories.addAssetPhoto(asset.id, photo);
      } catch {
        failedPhotoCount += 1;
      }
    }

    if (requestIdentity) this.printAttempts.delete(requestIdentity);
    return {
      printJobId: asset.printJobId,
      id: asset.id,
      title: asset.title,
      message:
        failedPhotoCount > 0
          ? t("assets.savedWithPhotoFailures", { title: asset.title, count: failedPhotoCount })
          : t("assets.savedNamed", { title: asset.title })
    };
  }
}
