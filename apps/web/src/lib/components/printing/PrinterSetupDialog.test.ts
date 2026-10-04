import {afterEach,expect,it} from 'vitest';
import {mount,tick,unmount} from 'svelte';
import PrinterSetupDialog from './PrinterSetupDialog.svelte';
import type {TextClipboard} from '$lib/ports/textClipboard';
class ControlledClipboard implements TextClipboard{
 content='';denied=false;pending?:Promise<void>;
 async write(text:string){await this.pending;if(this.denied)throw new Error('Denied');this.content=text;}
}
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);document.body.innerHTML='';});
async function settle(){for(let i=0;i<8;i++){await Promise.resolve();await tick();}}
it('keeps the safely quoted API command selectable when clipboard access is denied',async()=>{
 const clipboard=new ControlledClipboard();clipboard.denied=true;
 component=mount(PrinterSetupDialog,{target:document.body,props:{apiBaseUrl:'https://api.example.test',clipboard,onClose(){},onRestoreFocus(){}}});await settle();
 const name=document.getElementById('printer-computer-name') as HTMLInputElement;
 name.value="Garage's $(computer)";name.dispatchEvent(new Event('input',{bubbles:true}));await settle();
 const button=[...document.querySelectorAll('button')].find(button=>button.textContent?.trim()==='Copy command')!;
 button.click();await settle();expect(document.querySelector('[role="alert"]')?.textContent).toContain('Select and copy');
 const command=document.querySelector('code')!.textContent!;expect(command).toBe("./stuffstash --server 'https://api.example.test' connectors print register --name 'Garage'\\''s $(computer)'");
 clipboard.denied=false;button.click();await settle();expect(clipboard.content).toBe(command);expect(document.body.textContent).toContain('Command copied');
 let release!:()=>void;clipboard.pending=new Promise<void>(resolve=>{release=resolve;});button.click();await settle();
 name.value='Changed while copying';name.dispatchEvent(new Event('input',{bubbles:true}));await settle();release();await settle();
 expect(clipboard.content).toBe(command);expect(document.body.textContent).not.toContain('Command copied');
});
