import { colors } from '../theme/tokens';

type AssetNativeSheetOptions = {
  readonly contentStyle: { readonly backgroundColor: string };
  readonly headerShown: false;
  readonly presentation: 'formSheet';
  readonly gestureEnabled?: boolean;
  readonly sheetAllowedDetents: number[];
  readonly sheetCornerRadius: number;
  readonly sheetExpandsWhenScrolledToEdge: true;
  readonly sheetGrabberVisible: true;
  readonly sheetInitialDetentIndex: number;
  readonly sheetLargestUndimmedDetentIndex: 'none';
};

const baseAssetNativeSheetOptions = {
  contentStyle: { backgroundColor: colors.surface },
  headerShown: false,
  presentation: 'formSheet',
  sheetCornerRadius: 24,
  sheetExpandsWhenScrolledToEdge: true,
  sheetGrabberVisible: true,
  sheetInitialDetentIndex: 0,
  sheetLargestUndimmedDetentIndex: 'none'
} as const;

export const assetEditNativeSheetOptions: AssetNativeSheetOptions = {
  ...baseAssetNativeSheetOptions,
  gestureEnabled: false,
  sheetAllowedDetents: [0.56, 0.9]
};

export const assetMoveNativeSheetOptions: AssetNativeSheetOptions = {
  ...baseAssetNativeSheetOptions,
  sheetAllowedDetents: [0.62, 0.92]
};

export const assetMoveHereNativeSheetOptions: AssetNativeSheetOptions = {
  ...baseAssetNativeSheetOptions,
  sheetAllowedDetents: [0.6, 0.9]
};

export const assetAuditNativeSheetOptions: AssetNativeSheetOptions = {
  ...baseAssetNativeSheetOptions,
  sheetAllowedDetents: [0.58, 0.92]
};

export const assetCheckoutHistoryNativeSheetOptions: AssetNativeSheetOptions = {
  ...baseAssetNativeSheetOptions,
  sheetAllowedDetents: [0.58, 0.92]
};
