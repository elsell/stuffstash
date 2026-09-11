import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetDetailIdentitySection } from './AssetDetailIdentitySection';
import { toAssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { mapAsset } from '../../adapters/inventories/InventoryAssetMapping';

it.each([
 ['upcoming',true,'Expiring soon'],['expired',true,'Expired'],['expired',false,'Expiration tracking disabled'],['current',true,''],['missing',true,'']
] as const)('shows server expiration state %s with tracking %s', async(state,trackingEnabled,expected)=>{
 const harness=new MobileRenderHarness();
 const asset=toAssetDetailViewModel(mapAsset('Home',{id:'bottle',tenantId:'home',inventoryId:'main',title:'Bottle',kind:'item',lifecycleState:'active',description:'',parentAssetId:null,createdAt:'2026-09-01T00:00:00Z',updatedAt:'2026-09-01T00:00:00Z',customFields:{},tags:[],expiration:{date:'2028-02',precision:'month'},expirationContext:state==='missing'?undefined:{state,trackingEnabled,advanceDays:14,timezone:'America/New_York'}},[]));
 try {
  await harness.render(<AssetDetailIdentitySection asset={asset} isActionPending={false}/>);
  const text=harness.allText().join(' ');
  if(expected) expect(text).toContain(expected);
  else {expect(text).not.toContain('Expiring soon');expect(text).not.toContain('Expiration tracking disabled');}
 } finally{await harness.unmount();}
});
