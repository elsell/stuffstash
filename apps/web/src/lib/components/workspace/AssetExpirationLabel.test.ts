import {expect,it} from 'vitest';
import {mount,unmount} from 'svelte';
import AssetExpirationLabel from './AssetExpirationLabel.svelte';
it('shows date precision and omits absent expiration without an empty label',async()=>{
 for(const [expiration,expected] of [[{date:'2028-02',precision:'month'},'February 2028 (end of month)'],[{date:'2028-02-29',precision:'day'},'February 29, 2028'],[undefined,'']] as const){
  const target=document.createElement('div');document.body.append(target);
  const component=mount(AssetExpirationLabel,{target,props:{expiration,locale:'en-US'}});
  expect(target.textContent).toBe(expected?`Expiration: ${expected}`:'');
  await unmount(component);target.remove();
 }
});
