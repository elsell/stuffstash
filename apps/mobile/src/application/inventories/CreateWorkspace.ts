export type CreatedHousehold = { readonly id: string; readonly name: string; readonly canCreateInventory: boolean };
export type CreatedInventory = { readonly id: string; readonly tenantId: string; readonly name: string };
export interface WorkspaceCreationPort {
  canCreateInventory(tenantId: string): Promise<boolean>;
  createHousehold(name: string): Promise<CreatedHousehold>;
  createInventory(tenantId: string, name: string): Promise<CreatedInventory>;
}
export interface WorkspaceCreationObserver { created(): void }
/** Resource creation never changes the current connection or selected inventory. */
export class CreateWorkspace {
  constructor(private readonly port: WorkspaceCreationPort, private readonly observer: WorkspaceCreationObserver) {}
  async household(name: string): Promise<CreatedHousehold> {
    const result = await this.port.createHousehold(requiredName(name));
    this.observer.created();
    return result;
  }
  async inventory(tenantId: string, name: string): Promise<CreatedInventory> {
    const normalizedName = requiredName(name);
    if (!tenantId.trim() || !await this.port.canCreateInventory(tenantId)) throw new Error('You cannot create an inventory in this household.');
    const result = await this.port.createInventory(tenantId, normalizedName);
    this.observer.created();
    return result;
  }
}
function requiredName(name: string) {
  const value = name.trim();
  if (!value) throw new Error('Enter a name.');
  return value;
}
