import jsQR from 'jsqr';
import type {LabelCamera} from '$lib/ports/labels';
/** Only local camera pixels reach the decoder; scanned URLs are not fetched. */
export class BrowserLabelCamera implements LabelCamera {
  async start(video:HTMLVideoElement,onCode:(value:string)=>void,signal:AbortSignal):Promise<void> {
    const stream=await navigator.mediaDevices.getUserMedia({video:{facingMode:{ideal:'environment'},width:{ideal:960}},audio:false});
    let stopped=false;
    const hidden=()=>{if(document.hidden)stop();};
    const stop=()=>{stopped=true;stream.getTracks().forEach(track=>track.stop());if(video.srcObject===stream)video.srcObject=null;document.removeEventListener('visibilitychange',hidden);signal.removeEventListener('abort',stop);};
    if(signal.aborted){stop();return;}
    signal.addEventListener('abort',stop,{once:true});
    document.addEventListener('visibilitychange',hidden);
    video.srcObject=stream;
    try{await video.play();}catch(error){stop();signal.removeEventListener('abort',stop);throw error;}
    const canvas=document.createElement('canvas');const context=canvas.getContext('2d',{willReadFrequently:true});
    if(!context){stop();throw new Error('Camera unavailable');}
    let lastFrame=0;
    const decode=(now:number)=>{
      if(signal.aborted||stopped)return;
      if(now-lastFrame>=150&&video.videoWidth){
        lastFrame=now;canvas.width=Math.min(video.videoWidth,960);canvas.height=Math.round(video.videoHeight*canvas.width/video.videoWidth);
        context.drawImage(video,0,0,canvas.width,canvas.height);
        const pixels=context.getImageData(0,0,canvas.width,canvas.height);
        const code=jsQR(pixels.data,pixels.width,pixels.height,{inversionAttempts:'dontInvert'});
        if(code){stop();signal.removeEventListener('abort',stop);onCode(code.data);return;}
      }
      requestAnimationFrame(decode);
    };
    requestAnimationFrame(decode);
  }
}
