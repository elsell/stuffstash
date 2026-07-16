import type { MobileColorPalette } from '../theme/tokens';

export type AppNoticeTone = 'success' | 'info' | 'warning' | 'error';

export type AppNoticeInput = {
  readonly tone: AppNoticeTone;
  readonly title: string;
  readonly message?: string;
  readonly actionLabel?: string;
};

export type AppNoticePresentation = {
  readonly accessibilityLabel: string;
  readonly backgroundColor: string;
  readonly borderColor: string;
  readonly durationMs: number;
  readonly message?: string;
  readonly textColor: string;
  readonly title: string;
};

export function buildAppNoticePresentation(
  input: AppNoticeInput,
  colors: MobileColorPalette
): AppNoticePresentation {
  const title = input.title.trim();
  const message = input.message?.trim() || undefined;
  const palette = noticePalette(input.tone, colors);

  return {
    accessibilityLabel: [title, message].filter(Boolean).join('. '),
    backgroundColor: palette.backgroundColor,
    borderColor: palette.borderColor,
    durationMs: input.actionLabel ? 6500 : 4200,
    message,
    textColor: palette.textColor,
    title
  };
}

function noticePalette(tone: AppNoticeTone, colors: MobileColorPalette): {
  readonly backgroundColor: string;
  readonly borderColor: string;
  readonly textColor: string;
} {
  switch (tone) {
    case 'success':
      return {
        backgroundColor: colors.successSurface,
        borderColor: colors.successBorder,
        textColor: colors.text
      };
    case 'warning':
      return {
        backgroundColor: colors.warningSurface,
        borderColor: colors.warningBorder,
        textColor: colors.text
      };
    case 'error':
      return {
        backgroundColor: colors.dangerSurface,
        borderColor: colors.dangerBorder,
        textColor: colors.text
      };
    case 'info':
    default:
      return {
        backgroundColor: colors.elevatedSurface,
        borderColor: colors.border,
        textColor: colors.text
      };
  }
}
