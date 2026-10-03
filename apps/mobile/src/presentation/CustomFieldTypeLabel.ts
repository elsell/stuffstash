import { t } from './localization';
import type { CustomFieldType } from '../domain/customization/Customization';

export function customFieldTypeLabel(type: CustomFieldType): string {
  return t(`customization.fieldType.${type}`);
}
