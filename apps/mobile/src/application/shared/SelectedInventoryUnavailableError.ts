export class SelectedInventoryUnavailableError extends Error {
  constructor() { super('The selected Stuff Stash inventory is no longer available.'); }
}
