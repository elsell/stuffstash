import { t } from '../../presentation/localization';
export class SelectedInventoryUnavailableError extends Error {
  constructor() { super(t('mobile.SelectedInventoryUnavailableError.theSelectedStuffStashInventoryIsNoLongerAvailable')); }
}
