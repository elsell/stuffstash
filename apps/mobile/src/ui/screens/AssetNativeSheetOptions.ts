import { colors, type MobileColorPalette } from '../theme/tokens';

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

function baseAssetNativeSheetOptions(palette: MobileColorPalette) {
  return {
    contentStyle: { backgroundColor: palette.surface },
    headerShown: false,
    presentation: 'formSheet',
    sheetCornerRadius: 24,
    sheetExpandsWhenScrolledToEdge: true,
    sheetGrabberVisible: true,
    sheetInitialDetentIndex: 0,
    sheetLargestUndimmedDetentIndex: 'none'
  } as const;
}

export function createAssetNativeSheetOptions(palette: MobileColorPalette) {
  const baseOptions = baseAssetNativeSheetOptions(palette);
  return {
    edit: {
      ...baseOptions,
      gestureEnabled: false,
      sheetAllowedDetents: [0.56, 0.9]
    } satisfies AssetNativeSheetOptions,
    move: {
      ...baseOptions,
      sheetAllowedDetents: [0.62, 0.92]
    } satisfies AssetNativeSheetOptions,
    moveHere: {
      ...baseOptions,
      sheetAllowedDetents: [0.6, 0.9]
    } satisfies AssetNativeSheetOptions,
    checkoutHistory: {
      ...baseOptions,
      sheetAllowedDetents: [0.58, 0.92]
    } satisfies AssetNativeSheetOptions
  };
}

const defaultOptions = createAssetNativeSheetOptions(colors);

export const assetEditNativeSheetOptions: AssetNativeSheetOptions = {
  ...defaultOptions.edit
};

export const assetMoveNativeSheetOptions: AssetNativeSheetOptions = {
  ...defaultOptions.move
};

export const assetMoveHereNativeSheetOptions: AssetNativeSheetOptions = {
  ...defaultOptions.moveHere
};

export const assetCheckoutHistoryNativeSheetOptions: AssetNativeSheetOptions = {
  ...defaultOptions.checkoutHistory
};
