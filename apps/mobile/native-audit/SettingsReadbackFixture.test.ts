import { expect, it } from 'vitest';
import { mobileQueryKeys } from '../src/adapters/serverState/MobileQueryClient';
import { createSettingsReadback } from './SettingsReadbackFixture';

it('retains prior data after rejection and reads the successful retry through the collection query', async () => {
 const state=createSettingsReadback(); const context=await state.context.execute();
 await expect(state.tags.update(context,'tools',{displayName:'Emergency tools'})).rejects.toThrow('temporarily unavailable');
 expect((await state.query.tags(context)).items.map(tag=>tag.displayName)).toEqual(['Tools']);
 await state.tags.update(context,'tools',{displayName:'Emergency tools'});
 expect((await state.query.tags(context)).items).toEqual([expect.objectContaining({id:'tools',displayName:'Emergency tools'})]);
 const created=await state.tags.create(context,{displayName:'Camping'});
 expect((await state.query.tags(context)).items.map(tag=>tag.displayName)).toEqual(['Camping','Emergency tools']);
 await state.tags.archive(context,created.id);
 expect((await state.query.tags(context)).items.map(tag=>tag.displayName)).toEqual(['Emergency tools']);
 state.client.clear();
});
it('rejects unknown IDs and another inventory without changing stored tags', async()=>{
 const state=createSettingsReadback();const context=await state.context.execute();
 await expect(state.tags.update(context,'missing',{displayName:'Wrong'})).rejects.toThrow();
 await expect(state.tags.create({...context,inventoryId:'other'},{displayName:'Wrong'})).rejects.toThrow();
 expect((await state.query.tags(context)).items.map(tag=>tag.displayName)).toEqual(['Tools']);state.client.clear();
});

it('invalidates cached collection reads only after a successful command',async()=>{
 const state=createSettingsReadback();const context=await state.context.execute();
 const queryKey=mobileQueryKeys.customization('settings-readback',context.tenantId,context.inventoryId,'inventory','tag','active');
 const options={queryKey,queryFn:()=>state.query.tags(context),staleTime:Infinity};
 await state.client.fetchQuery(options);
 await expect(state.tags.update(context,'tools',{displayName:'Emergency tools'})).rejects.toThrow();
 expect(state.client.getQueryState(queryKey)?.isInvalidated).toBe(false);
 await state.tags.update(context,'tools',{displayName:'Emergency tools'});
 expect(state.client.getQueryState(queryKey)?.isInvalidated).toBe(true);
 expect((await state.client.fetchQuery(options)).items.map(tag=>tag.displayName)).toEqual(['Emergency tools']);
 state.client.clear();
});
