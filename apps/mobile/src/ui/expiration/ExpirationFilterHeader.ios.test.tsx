import { expect, it } from 'vitest';
import { expirationFilterHeaderOptions } from './ExpirationFilterHeader.ios';
it('installs a system filter button with a non-color active indication and action',()=>{
 let opened=0;
 for(const active of [false,true]) {
  const options=expirationFilterHeaderOptions({active,onPress:()=>{opened++}});
  const item=options.unstable_headerRightItems?.({canGoBack:true})[0];
  expect(item?.type).toBe('button');
  if (item?.type !== 'button') throw new Error('Expected native filter button');
  expect(item?.accessibilityLabel).toBe(active?'Filter expiration items, filters active':'Filter expiration items');
  expect(item?.icon).toEqual({type:'sfSymbol',name:active?'line.3.horizontal.decrease.circle.fill':'line.3.horizontal.decrease.circle'});
  if(item?.type==='button') item.onPress?.();
 }
 expect(opened).toBe(2);
});
