import { createTranslator, en } from '@stuff-stash/localization';

// Build-time verification override. Never mutate a locale shared with another request.
export const localization = createTranslator(en, { locale: import.meta.env.VITE_STUFF_STASH_UI_LOCALE });
export const t = localization.message;
