import { t } from '../../presentation/localization';
import { Platform } from 'react-native';
import { colors, type MobileColorPalette } from '../theme/tokens';

type AssetNativeSheetOptions = {
  readonly contentStyle: { readonly backgroundColor: string };
  readonly headerShown: boolean;
  readonly title?: string;
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

export function createAssetNativeSheetOptions(palette: MobileColorPalette, platform: string = Platform.OS) {
  const baseOptions = baseAssetNativeSheetOptions(palette);
  return {
    selection: {
      contentStyle: { backgroundColor: palette.surface }, headerShown: true,
      presentation: platform === 'ios' ? 'fullScreenModal' as const : 'card' as const
    },
    add: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.background },
      presentation: 'card' as const, headerShown: true, title: t('mobile.AssetNativeSheetOptions.addItem')
    } : {
      contentStyle: { backgroundColor: palette.background },
      presentation: 'formSheet' as const, headerShown: true, title: t('mobile.AssetNativeSheetOptions.addItem'),
      sheetAllowedDetents: [1], sheetCornerRadius: 24, sheetGrabberVisible: true
    },
    filters: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.background },
      presentation: 'card' as const, headerShown: true, title: t('mobile.AssetNativeSheetOptions.filters')
    } : {
      contentStyle: { backgroundColor: palette.surface },
      presentation: 'modal' as const,
      headerShown: true,
      title: t('mobile.AssetNativeSheetOptions.filters')
    },
    edit: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.surface }, presentation: 'card' as const,
      headerShown: true, gestureEnabled: false
    } : {
      ...baseOptions, headerShown: true, title: t('mobile.AssetNativeSheetOptions.editAsset'), gestureEnabled: false, sheetAllowedDetents: [1]
    } satisfies AssetNativeSheetOptions,
    move: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.surface }, presentation: 'card' as const, headerShown: true
    } : {
      ...baseOptions, headerShown: true, title: t('mobile.AssetNativeSheetOptions.moveAsset'), sheetAllowedDetents: [1]
    } satisfies AssetNativeSheetOptions,
    moveHere: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.surface }, presentation: 'card' as const, headerShown: true
    } : {
      ...baseOptions, headerShown: true, title: t('mobile.AssetNativeSheetOptions.moveSomethingHere'), sheetAllowedDetents: [1]
    } satisfies AssetNativeSheetOptions,
    checkoutHistory: platform === 'android' ? {
      contentStyle: { backgroundColor: palette.surface }, presentation: 'card' as const,
      headerShown: true, title: t('mobile.AssetNativeSheetOptions.checkoutHistory')
    } : {
      ...baseOptions,
      headerShown: true,
      title: t('mobile.AssetNativeSheetOptions.checkoutHistory'),
      sheetAllowedDetents: [0.58, 0.92]
    } satisfies AssetNativeSheetOptions
  };
}

const defaultOptions = createAssetNativeSheetOptions(colors);

export const assetEditNativeSheetOptions = {
  ...defaultOptions.edit
};

export const assetMoveNativeSheetOptions = {
  ...defaultOptions.move
};

export const assetMoveHereNativeSheetOptions = {
  ...defaultOptions.moveHere
};

export const assetCheckoutHistoryNativeSheetOptions = {
  ...defaultOptions.checkoutHistory
};
