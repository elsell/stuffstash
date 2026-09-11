import React from 'react';
import {expect,it} from 'vitest';
import {MobileRenderHarness} from '../../test-support/render';
import {TimeZonePicker} from './TimeZonePicker';
it('searches readable cities, keeps the saved choice and handles a failed save',async()=>{
 const h=new MobileRenderHarness();const saved:string[]=[];
 try{await h.render(<TimeZonePicker value="America/New_York" onChange={async zone=>{saved.push(zone);throw new Error('private');}}/>);
 expect(h.byLabel('New York · America')!.props.accessibilityState.checked).toBe(true);
 await h.changeText(h.byLabel('Search time zones'),'NoSuchCity');expect(h.allText()).toContain('No matching time zones.');
 await h.changeText(h.byLabel('Search time zones'),'UTC');await h.press(h.byLabel('UTC'));
 expect(saved).toEqual(['UTC']);expect(h.allText()).toContain('Could not save the time zone. Try again.');
 expect(h.byLabel('UTC')!.props.accessibilityState.checked).toBe(false);
 }finally{await h.unmount();}
});
