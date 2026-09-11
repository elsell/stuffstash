import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ReminderTimingEditor } from './ReminderTimingEditor';
const policy={enabled:true,upcoming:true,expired:true,advanceDays:30};
it('saves a preset while preserving independent expired reminders',async()=>{
 const h=new MobileRenderHarness();const saved:unknown[]=[];let closed=false;
 try{await h.render(<ReminderTimingEditor policy={policy} onSave={async p=>{saved.push(p);}} onDone={()=>{closed=true;}}/>);
 await h.press(h.byLabel('1 day before'));expect(saved).toEqual([{...policy,advanceDays:1}]);expect(closed).toBe(true);
 }finally{await h.unmount();}
});
it('keeps custom input after save failure and never saves invalid or cancelled input',async()=>{
 const h=new MobileRenderHarness();const saved:unknown[]=[];let closed=false;let fail=true;
 try{await h.render(<ReminderTimingEditor policy={policy} onSave={async p=>{saved.push(p);if(fail)throw new Error('private');}} onDone={()=>{closed=true;}}/>);
 await h.press(h.byLabel('Custom days'));
 await h.changeText(h.byLabel('Days before expiration'),'1.5');
 const done=()=>h.byLabel('Save reminder days')!;
 expect(done().props.disabled).toBe(true);
 await h.changeText(h.byLabel('Days before expiration'),'12');
 await h.run(()=>done().props.onPress());expect(closed).toBe(false);expect(h.byLabel('Days before expiration')!.props.value).toBe('12');
 fail=false;await h.run(()=>done().props.onPress());expect(closed).toBe(true);expect(saved).toEqual([{...policy,advanceDays:12},{...policy,advanceDays:12}]);
 }finally{await h.unmount();}
});
it('discards unsaved custom days when leaving and restores clean refreshed values',async()=>{
 const h=new MobileRenderHarness();let writes=0;
 const render=(advanceDays:number)=>h.render(<ReminderTimingEditor policy={{...policy,advanceDays}} onSave={async()=>{writes++;}} onDone={()=>{}}/>);
 try{await render(30);await h.press(h.byLabel('Custom days'));await h.changeText(h.byLabel('Days before expiration'),'19');await render(7);expect(h.byLabel('Days before expiration')!.props.value).toBe('19');expect(writes).toBe(0);
 }finally{await h.unmount();}
});
it('shows the attempted preset after failure and supports retrying it',async()=>{
 const h=new MobileRenderHarness();let attempts=0;
 try{await h.render(<ReminderTimingEditor policy={policy} onSave={async()=>{attempts++;throw new Error('failed');}} onDone={()=>{}}/>);
 await h.press(h.byLabel('7 days before'));expect(h.byLabel('7 days before')!.props.accessibilityState.checked).toBe(true);
 await h.press(h.byLabel('7 days before'));expect(attempts).toBe(2);
 }finally{await h.unmount();}
});
it('shows Off even when the retained threshold is custom',async()=>{
 const h=new MobileRenderHarness();
 try{await h.render(<ReminderTimingEditor policy={{...policy,upcoming:false,advanceDays:12}} onSave={async()=>{}} onDone={()=>{}}/>);
 expect(h.byLabel('Off')!.props.accessibilityState.checked).toBe(true);expect(h.byLabel('Days before expiration')).toBeUndefined();
 }finally{await h.unmount();}
});
