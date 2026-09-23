import { createContext, useContext, useState, type PropsWithChildren } from 'react';
import { Stack, router, useLocalSearchParams, type Href } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { CustomizationAccessPolicy } from '../src/application/customization/CustomizationAccess';
import { CustomizationContextQuery } from '../src/application/customization/CustomizationContextQuery';
import { noCustomizationObservability } from '../src/application/customization/CustomizationObservability';
import { CustomizationCollectionQuery } from '../src/application/customization/CustomizationQueries';
import type { CustomizationContext, CustomizationRepository } from '../src/application/customization/CustomizationRepository';
import { ManageTags } from '../src/application/customization/ManageTags';
import { ManageCustomFields } from '../src/application/customization/ManageCustomFields';
import { ManageCustomAssetTypes } from '../src/application/customization/ManageCustomAssetTypes';
import type { AssetTagDefinition } from '../src/domain/customization/Customization';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { CustomizationEditorScreen } from '../src/ui/screens/CustomizationEditorScreen';
import { CustomizationCollectionScreen } from '../src/ui/screens/CustomizationCollectionScreen';

const scope = { tenantId: 'readback-household', inventoryId: 'readback-inventory' };
const unavailable = async (): Promise<never> => { throw new Error('Outside the tag readback journey'); };

export function createSettingsReadback() {
 let sequence=0;let rejectUpdate=true;
 const records=new Map<string,AssetTagDefinition>([['tools',{kind:'tag',id:'tools',key:'tools',displayName:'Tools'}]]);
 function verify(context:CustomizationContext) {
  if(context.tenantId!==scope.tenantId||context.inventoryId!==scope.inventoryId)throw new Error('Outside the fixture inventory');
 }
 const repository:CustomizationRepository={
  async listTags(context){verify(context);return {items:[...records.values()]};},
  async createTag(context,input){verify(context);const id=`created-${++sequence}`;const tag:AssetTagDefinition={kind:'tag',id,key:id,...input};records.set(id,tag);return tag;},
  async updateTag(context,id,input){verify(context);const tag=records.get(id);if(!tag)throw new Error('Unknown tag');
   if(rejectUpdate){rejectUpdate=false;throw new Error('Settings temporarily unavailable');}
   const updated={...tag,...input};records.set(id,updated);return updated;},
  async archiveTag(context,id){verify(context);if(!records.delete(id))throw new Error('Unknown tag');},
  listFields:unavailable,listAssetTypes:unavailable,
  createField:unavailable,updateField:unavailable,archiveField:unavailable,restoreField:unavailable,deleteField:unavailable,
  createAssetType:unavailable,updateAssetType:unavailable,archiveAssetType:unavailable,restoreAssetType:unavailable,deleteAssetType:unavailable,
 };
 return {
  client:createMobileQueryClient(),policy:new CustomizationAccessPolicy(noCustomizationObservability),
  context:new CustomizationContextQuery({getSelectedScope:async()=>({tenant:{id:scope.tenantId,name:'Audit household',permissions:['configure']},inventory:{id:scope.inventoryId,name:'Audit inventory',permissions:['view','edit_asset']}})}),
  query:new CustomizationCollectionQuery(repository),tags:new ManageTags(repository,noCustomizationObservability),
  fields:new ManageCustomFields(repository,noCustomizationObservability),types:new ManageCustomAssetTypes(repository,noCustomizationObservability),
 };
}
const JourneyContext=createContext<ReturnType<typeof createSettingsReadback>|undefined>(undefined);
const loadScope=async()=>scope;
export function SettingsReadbackProvider({children}:PropsWithChildren){
 const [state]=useState(createSettingsReadback);
 return <MobileServerStateProvider client={state.client} scopeId="settings-readback" loadInventoryScope={loadScope}>
  <JourneyContext.Provider value={state}>{children}</JourneyContext.Provider>
 </MobileServerStateProvider>;
}
export function SettingsReadbackFixture(){
 const state=useContext(JourneyContext);if(!state)throw new Error('Settings journey provider missing');
 const {edit,create}=useLocalSearchParams<{edit?:string;create?:string}>();
 const editing=!!edit||create==='true';
 return <>
  <Stack.Screen options={{title:editing?create==='true'?'New Tag':'Edit Tag':'Tags'}} />
  {editing?<CustomizationEditorScreen accessPolicy={state.policy} contextQuery={state.context} query={state.query}
   kind="tag" scope="inventory" mode={create==='true'?'create':'edit'} resourceId={edit}
   manageTags={state.tags} manageFields={state.fields} manageAssetTypes={state.types} onDone={()=>router.back()} />
   :<CustomizationCollectionScreen accessPolicy={state.policy} contextQuery={state.context} query={state.query} kind="tag" scope="inventory"
    onAdd={()=>router.push({pathname:'/audit-settings-readback',params:{create:'true'}} as Href)}
    onOpen={row=>router.push({pathname:'/audit-settings-readback',params:{edit:row.id}} as Href)} />}
 </>;
}
