export type SelectedAssetPhoto = {
  readonly id: string;
  readonly uri: string;
  readonly fileName: string;
  readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp';
  readonly contentBase64: string;
};

export interface PhotoSelectionProvider {
  selectPhotos(existingCount: number): Promise<readonly SelectedAssetPhoto[]>;
}

export class PhotoSelectionQuery {
  constructor(private readonly provider: PhotoSelectionProvider) {}

  async execute(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
    return this.provider.selectPhotos(existingCount);
  }
}
