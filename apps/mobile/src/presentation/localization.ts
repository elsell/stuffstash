import { createTranslator, en } from '@stuff-stash/localization';

// Explicit pseudolocale builds are for verification; normal builds use the device locale.
export const localization = createTranslator(en, { locale: process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE });
export const t = localization.message;
