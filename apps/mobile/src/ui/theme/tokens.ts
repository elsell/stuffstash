export const lightPalette = {
  brandCharcoal: '#303A41',
  brandCharcoalDeep: '#243038',
  brandDustyBlue: '#6B90AA',
  brandDustyBlueSoft: '#E8F0F5',
  brandAmber: '#F5AB4B',
  background: '#F7FAFB',
  surface: '#ffffff',
  elevatedSurface: '#ffffff',
  surfaceMuted: '#E8F0F5',
  text: '#243038',
  textMuted: '#52616B',
  border: '#D9E1E6',
  controlBorder: '#6F7E88',
  accent: '#6B90AA',
  accentStrong: '#303A41',
  selected: '#E8F0F5',
  action: '#0066CC',
  actionPressed: '#004F9F',
  focusRing: '#8A4F00',
  scrim: 'rgba(0, 0, 0, 0.68)',
  onScrim: '#ffffff',
  warningSurface: '#FFF3DF',
  warningBorder: '#8A4F00',
  warning: '#8A4F00',
  successSurface: '#E7F5EE',
  successBorder: '#237A57',
  success: '#237A57',
  dangerSurface: '#FDECEC',
  dangerBorder: '#C03535',
  danger: '#C03535',
  onAction: '#ffffff'
} as const;

export const darkPalette: MobileColorPalette = {
  brandCharcoal: '#D8E0E5',
  brandCharcoalDeep: '#EDF2F5',
  brandDustyBlue: '#8EB3CC',
  brandDustyBlueSoft: '#22333F',
  brandAmber: '#F5B95E',
  background: '#111416',
  surface: '#1C2023',
  elevatedSurface: '#252A2E',
  surfaceMuted: '#27343C',
  text: '#F4F7F8',
  textMuted: '#B8C4CB',
  border: '#2B3338',
  controlBorder: '#98A6AF',
  accent: '#8EB3CC',
  accentStrong: '#E4EEF4',
  selected: '#2B414F',
  action: '#72BCFF',
  actionPressed: '#9DCEFF',
  focusRing: '#FFC46B',
  scrim: 'rgba(0, 0, 0, 0.72)',
  onScrim: '#ffffff',
  warningSurface: '#3B2B13',
  warningBorder: '#FFD28A',
  warning: '#FFD28A',
  successSurface: '#17392D',
  successBorder: '#7DDDB3',
  success: '#7DDDB3',
  dangerSurface: '#452126',
  dangerBorder: '#FF9B9B',
  danger: '#FF9B9B',
  onAction: '#071A2A'
};

export const lightHighContrastPalette: MobileColorPalette = {
  ...lightPalette,
  border: '#9AA7AF',
  controlBorder: '#303A41',
  focusRing: '#6B3900'
};

export const darkHighContrastPalette: MobileColorPalette = {
  ...darkPalette,
  border: '#68757D',
  controlBorder: '#F4F7F8',
  focusRing: '#FFE0A3'
};

export type MobileColorPalette = {
  readonly [Key in keyof typeof lightPalette]: string;
};

export const colors = lightPalette;

export function mobileColorPalette(
  colorScheme: 'light' | 'dark' | 'unspecified' | null | undefined,
  increasedContrast = false
): MobileColorPalette {
  if (colorScheme === 'dark') {
    return increasedContrast ? darkHighContrastPalette : darkPalette;
  }
  return increasedContrast ? lightHighContrastPalette : lightPalette;
}

export const spacing = {
  xs: 6,
  sm: 10,
  md: 16,
  lg: 24,
  xl: 32
} as const;

export const radius = {
  sm: 6,
  md: 8,
  lg: 16
} as const;

export const typography = {
  wordmarkWeight: '600',
  titleWeight: '800',
  bodyWeight: '400'
} as const;
