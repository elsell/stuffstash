import type {
  CreatedInventoryAccessInvitation,
  InventoryAccessGrant,
  InventoryAccessInvitation,
  InventoryAccessRelationship,
  InvitationStatusFilter
} from '$lib/domain/inventory';
import type { Pagination } from './pagination';

export interface InventoryAccessPage<T> {
  items: T[];
  pagination: Pagination;
}

export interface InventoryAccessRepository {
  listInventoryAccessGrants(
    tenantId: string,
    inventoryId: string,
    cursor?: string
  ): Promise<InventoryAccessPage<InventoryAccessGrant>>;
  grantInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<InventoryAccessGrant>;
  revokeInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<void>;
  listInventoryAccessInvitations(
    tenantId: string,
    inventoryId: string,
    status: InvitationStatusFilter,
    cursor?: string
  ): Promise<InventoryAccessPage<InventoryAccessInvitation>>;
  createInventoryAccessInvitation(
    tenantId: string,
    inventoryId: string,
    email: string,
    relationship: InventoryAccessRelationship
  ): Promise<CreatedInventoryAccessInvitation>;
  updateInventoryAccessInvitationExpiration(
    tenantId: string,
    inventoryId: string,
    invitationId: string,
    expiresAt: string
  ): Promise<InventoryAccessInvitation>;
  cancelInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void>;
  deleteInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void>;
}
