import type { SelectedAssetPhoto } from './PhotoSelectionQuery';
import type { AssetKind } from '../../domain/assets/AssetSummary';

export type AddAssetDraftParent = {
  readonly id: string;
  readonly title: string;
  readonly kind: AssetKind;
  readonly selectionHint: string;
  readonly subtitle: string;
  readonly pathLabel: string;
  readonly willPromoteToContainer: boolean;
};

export type AddAssetDraft = {
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly parentQuery: string;
  readonly selectedPhotos: readonly SelectedAssetPhoto[];
  readonly showDetails: boolean;
  readonly lastParent?: AddAssetDraftParent;
};

export type AddAssetDraftContext = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly principalId: string;
};

export interface AddAssetDraftStore {
  load(context: AddAssetDraftContext): AddAssetDraft | undefined;
  save(context: AddAssetDraftContext, draft: AddAssetDraft): void;
}

export class InMemoryAddAssetDraftStore implements AddAssetDraftStore {
  private readonly drafts = new Map<string, AddAssetDraft>();

  constructor(private readonly serviceScopeId: string) {}

  load(context: AddAssetDraftContext): AddAssetDraft | undefined {
    return this.drafts.get(this.key(context));
  }

  save(context: AddAssetDraftContext, draft: AddAssetDraft): void {
    this.drafts.set(this.key(context), draft);
  }

  private key(context: AddAssetDraftContext): string {
    return [
      this.serviceScopeId,
      context.principalId,
      context.tenantId,
      context.inventoryId
    ].join(':');
  }
}
