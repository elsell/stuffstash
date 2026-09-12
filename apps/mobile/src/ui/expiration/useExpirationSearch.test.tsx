import React from 'react';
import { expect, it, vi } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useExpirationSearch } from './useExpirationSearch';
it('debounces typing, applies clear immediately and restores external route changes', async () => {
 vi.useFakeTimers(); const harness=new MobileRenderHarness(); const values:string[]=[];
 let search: ReturnType<typeof useExpirationSearch>;
 function Surface({query='',prefix=''}:{query?:string;prefix?:string}) { search=useExpirationSearch(query,value=>values.push(prefix+value)); return null; }
 try {
  await harness.render(<Surface />);
  await harness.run(()=>search.change('H')); await harness.run(()=>search.change('Hey'));
  await harness.run(()=>vi.advanceTimersByTime(299)); expect(values).toEqual([]);
  await harness.run(()=>vi.advanceTimersByTime(1)); expect(values).toEqual(['Hey']);
  await harness.render(<Surface query="Hey" prefix="new:" />);
  await harness.run(()=>search.change('other')); await harness.run(()=>search.clear());
  await harness.run(()=>vi.advanceTimersByTime(300)); expect(values).toEqual(['Hey','new:']);
  await harness.run(()=>search.change('pending'));
  await harness.render(<Surface query="restored" prefix="new:" />);
  await harness.run(()=>vi.advanceTimersByTime(300)); expect(values).toEqual(['Hey','new:']);
  await harness.run(()=>search.submit('restored now')); expect(values.at(-1)).toBe('new:restored now');
  await harness.run(()=>search.change('unmounted')); await harness.unmount();
  await harness.run(()=>vi.advanceTimersByTime(300)); expect(values.at(-1)).toBe('new:restored now');
 } finally { await harness.unmount(); vi.useRealTimers(); }
});
