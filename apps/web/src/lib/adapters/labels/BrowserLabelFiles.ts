import type {LabelFileDelivery} from '$lib/ports/labels';
export class BrowserLabelFiles implements LabelFileDelivery {
  save(content:Blob,format:'png'|'pdf'):void {
    const url=URL.createObjectURL(content);const link=document.createElement('a');
    link.href=url;link.download=`stuff-stash-label.${format}`;link.click();setTimeout(()=>URL.revokeObjectURL(url),60000);
  }
  preparePrint(){
    // Reserve the window within the user gesture, before the authenticated fetch.
    const opened=window.open('about:blank','_blank');
    if(!opened)throw new Error('Print window unavailable');
    opened.opener=null;
    return {
      show(content:Blob){const url=URL.createObjectURL(content);opened.location.replace(url);setTimeout(()=>URL.revokeObjectURL(url),300000);},
      close(){opened.close();},
    };
  }
}
