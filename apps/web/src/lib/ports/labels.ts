import type {LabelScope,LabelReference,LabelDestination,LabelChoice,LabelCatalog,LabelArtifact} from '$lib/domain/label';
export interface LabelRepository {
  parse(value:string):LabelReference;
  resolve(reference:LabelReference,signal:AbortSignal):Promise<LabelDestination>;
  catalog(scope:LabelScope,signal:AbortSignal):Promise<LabelCatalog>;
  render(scope:LabelScope,choice:LabelChoice,format:'png'|'pdf',signal:AbortSignal):Promise<LabelArtifact>;
}
export interface LabelFileDelivery {save(content:Blob,format:'png'|'pdf'):void;preparePrint():{show(content:Blob):void;close():void}}
export interface LabelCamera {start(video:HTMLVideoElement,onCode:(value:string)=>void,signal:AbortSignal):Promise<void>}
export interface LabelWorkspace {repository:LabelRepository;files:LabelFileDelivery;camera:LabelCamera}
export const labelWorkspaceContext=Symbol('labelWorkspace');
