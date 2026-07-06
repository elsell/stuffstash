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
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure', 'view_import_job', 'create_import_job']
      }
    }
  ],
  customAssetTypes: [
    {
      id: 'type-document',
      tenantId: 'tenant-home',
      inventoryId: null,
      scope: 'tenant',
      key: 'document',
      displayName: 'Document',
      description: 'Passports, records, and household paperwork.',
      lifecycleState: 'active'
    },
    {
      id: 'type-garden-supply',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      scope: 'inventory',
      key: 'garden-supply',
      displayName: 'Garden supply',
      description: 'Fertilizer, soil amendments, and outdoor consumables.',
      lifecycleState: 'active'
    }
  ],
  customFieldDefinitions: [
    {
      id: 'field-expiration-date',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      scope: 'inventory',
      key: 'expiration-date',
      displayName: 'Expiration date',
      type: 'date',
      enumOptions: [],
      applicability: 'custom_asset_types',
      customAssetTypeIds: ['type-garden-supply'],
      lifecycleState: 'active'
    },
    {
      id: 'field-storage-note',
      tenantId: 'tenant-home',
      inventoryId: null,
      scope: 'tenant',
      key: 'storage-note',
      displayName: 'Storage note',
      type: 'text',
      enumOptions: [],
      applicability: 'all_assets',
      customAssetTypeIds: [],
      lifecycleState: 'active'
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
      customFields: {},
      photo: {
        id: 'photo-garage',
        assetId: 'asset-garage',
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
      lifecycleState: 'active',
      customFields: {}
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
      customFields: { 'storage-note': 'Middle shelf' },
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
      customAssetTypeId: 'type-garden-supply',
      customFields: { 'expiration-date': '2026-09-01' },
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
      customAssetTypeId: 'type-document',
      customFields: { 'storage-note': 'Blue folder' },
      customAssetTypeLabel: 'Document'
    }
  ]
};
