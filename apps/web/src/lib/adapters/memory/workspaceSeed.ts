import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';

export const workspaceSeed: WorkspaceSeed = {
  principal: {
    id: 'local-user',
    email: 'demo@stuffstash.local'
  },
  tenants: [
    {
      id: 'tenant-home',
      name: 'Home',
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_inventory', 'configure']
      }
    }
  ],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure']
      }
    }
  ],
  assets: [
    {
      id: 'asset-garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: 'Tools, outdoor supplies, and seasonal storage.',
      parentAssetId: null,
      lifecycleState: 'active',
      photo: {
        id: 'photo-garage',
        url: 'https://images.unsplash.com/photo-1585128792020-803d29415281?auto=format&fit=crop&w=900&q=70',
        alt: 'Garage shelving'
      }
    },
    {
      id: 'asset-hall-closet',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Hall closet',
      description: 'Medicine, documents, and household backups.',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-toolbox',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'container',
      title: 'Red toolbox',
      description: 'Socket set and drill bits.',
      parentAssetId: 'asset-garage',
      lifecycleState: 'active',
      customAssetTypeLabel: 'Tool storage'
    },
    {
      id: 'asset-fertilizer',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Tomato fertilizer',
      description: 'Half-full bag for raised beds.',
      parentAssetId: 'asset-garage',
      lifecycleState: 'active',
      customAssetTypeLabel: 'Garden supply'
    },
    {
      id: 'asset-passports',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Passports',
      description: 'Family documents in blue folder.',
      parentAssetId: 'asset-hall-closet',
      lifecycleState: 'active',
      customAssetTypeLabel: 'Document'
    }
  ]
};
