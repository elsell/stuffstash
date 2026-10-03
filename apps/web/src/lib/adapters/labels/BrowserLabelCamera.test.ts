import {expect,it} from 'vitest';
import {BrowserLabelCamera} from './BrowserLabelCamera';
class CameraStreamFake {
  stopped=false;
  getTracks(){return [{stop:()=>{this.stopped=true;}}];}
}
it('late camera permission cannot detach a newer camera session',async()=>{
  const original=Object.getOwnPropertyDescriptor(navigator,'mediaDevices');
  const originalContext=HTMLCanvasElement.prototype.getContext;
  // No frames arrive while permission races; a controlled drawing surface keeps
  // this test focused on ownership of live camera streams.
  HTMLCanvasElement.prototype.getContext=(()=>({drawImage(){},getImageData(){return {data:new Uint8ClampedArray(4),width:1,height:1};}})) as unknown as typeof originalContext;
  const waiting:((stream:MediaStream)=>void)[]=[];
  Object.defineProperty(navigator,'mediaDevices',{configurable:true,value:{getUserMedia:()=>new Promise<MediaStream>(resolve=>waiting.push(resolve))}});
  const video=document.createElement('video');video.play=async()=>{};
  const old=new AbortController(),current=new AbortController();
  const oldStream=new CameraStreamFake(),newStream=new CameraStreamFake();
  try{
    const camera=new BrowserLabelCamera();const first=camera.start(video,()=>{},old.signal);
    old.abort();const second=camera.start(video,()=>{},current.signal);
    waiting[1](newStream as unknown as MediaStream);await second;
    waiting[0](oldStream as unknown as MediaStream);await first;
    expect(oldStream.stopped).toBe(true);expect(newStream.stopped).toBe(false);
    expect(video.srcObject).toBe(newStream);
    current.abort();expect(newStream.stopped).toBe(true);expect(video.srcObject).toBeNull();
  }finally{old.abort();current.abort();HTMLCanvasElement.prototype.getContext=originalContext;if(original)Object.defineProperty(navigator,'mediaDevices',original);else Reflect.deleteProperty(navigator,'mediaDevices');}
});
