import React from 'react';
import {expect,it} from 'vitest';
import {MobileRenderHarness} from '../../test-support/render';
import {navigationOptions,resetNavigation} from '../../test-support/navigation';
import {TimeZonePicker} from './TimeZonePicker';
it('searches readable cities, keeps the saved choice and handles a failed save',async()=>{
 resetNavigation();const h=new MobileRenderHarness();const saved:string[]=[];
 try{await h.render(<TimeZonePicker value="America/New_York" onChange={async zone=>{saved.push(zone);throw new Error('private');}}/>);
 expect(h.byLabel('New York · America')!.props.accessibilityState.checked).toBe(true);
 const search=()=> (navigationOptions().at(-1) as {headerSearchBarOptions:{onChangeText:(e:{nativeEvent:{text:string}})=>void;onCancelButtonPress:()=>void}}).headerSearchBarOptions;
 expect(h.allByType('TextInput')).toHaveLength(0);
 await h.run(()=>search().onChangeText({nativeEvent:{text:'NoSuchCity'}}));expect(h.allText()).toContain('No matching time zones.');
 await h.run(()=>search().onChangeText({nativeEvent:{text:'UTC'}}));await h.press(h.byLabel('UTC'));
 expect(saved).toEqual(['UTC']);expect(h.allText()).toContain('Could not save the time zone. Try again.');
 expect(h.byLabel('UTC')!.props.accessibilityState.checked).toBe(false);
 await h.run(()=>search().onCancelButtonPress());
 expect(h.byLabel('New York · America')!.props.accessibilityState.checked).toBe(true);
 expect(saved).toEqual(['UTC']);
 }finally{await h.unmount();}
});
