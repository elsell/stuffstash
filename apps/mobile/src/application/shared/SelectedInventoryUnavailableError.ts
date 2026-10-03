import { CatalogRecoveryError } from './CatalogRecoveryError';
export class SelectedInventoryUnavailableError extends CatalogRecoveryError {
  constructor() { super('mobile.SelectedInventoryUnavailableError.theSelectedStuffStashInventoryIsNoLongerAvailable'); }
}
