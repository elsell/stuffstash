export type AssetActionCompletion = {
  readonly assetId: string;
  readonly action: 'edit' | 'move';
  readonly message: string;
};

const completions = new Map<string, AssetActionCompletion>();

export function recordAssetActionCompletion(completion: AssetActionCompletion): void {
  completions.set(completion.assetId, completion);
}

export function consumeAssetActionCompletion(assetId: string): AssetActionCompletion | undefined {
  const completion = completions.get(assetId);
  completions.delete(assetId);
  return completion;
}
