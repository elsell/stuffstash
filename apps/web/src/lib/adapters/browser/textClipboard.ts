import type {TextClipboard} from '$lib/ports/textClipboard';
export const browserTextClipboard:TextClipboard={
 async write(text){
  if(!navigator.clipboard?.writeText)throw new Error('Clipboard unavailable');
  await navigator.clipboard.writeText(text);
 }
};
