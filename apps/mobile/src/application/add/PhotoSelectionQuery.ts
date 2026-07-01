export type SelectedAssetPhoto = {
  readonly id: string;
  readonly uri: string;
  readonly fileName: string;
  readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp';
  readonly contentBase64?: string;
  readonly sizeBytes: number;
};

export interface PhotoSelectionProvider {
  selectFromLibrary(existingCount: number): Promise<readonly SelectedAssetPhoto[]>;
  captureFromCamera(existingCount: number): Promise<readonly SelectedAssetPhoto[]>;
}

export class PhotoSelectionQuery {
  constructor(private readonly provider: PhotoSelectionProvider) {}

  async selectFromLibrary(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
    return this.provider.selectFromLibrary(existingCount);
  }

  async captureFromCamera(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
    return this.provider.captureFromCamera(existingCount);
  }
}
