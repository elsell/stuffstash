import type { createTranslator, en } from '@stuff-stash/localization';
import { localization } from './localization';

export function timestampLabel(value: string, translator: ReturnType<typeof createTranslator<typeof en>> = localization): string {
  return new Date(value).toLocaleString(translator.locale);
}
