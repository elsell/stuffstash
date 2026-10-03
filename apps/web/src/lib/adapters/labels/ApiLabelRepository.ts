import {LabelsClient,parseLabelLink,type LabelMedia} from '@stuff-stash/api-client';
import type {LabelRepository} from '$lib/ports/labels';
import type {LabelReference,LabelScope,LabelChoice} from '$lib/domain/label';
export class ApiLabelRepository implements LabelRepository {
  private media=new Map<string,LabelMedia>();
  constructor(private readonly client:LabelsClient){}
  parse(value:string){return parseLabelLink(value);}
  async resolve(reference:LabelReference,signal:AbortSignal){
    const instance=await this.client.instance(signal);
    if(instance.instanceId!==reference.instanceId)throw new Error('wrong_instance');
    const result=await this.client.resolve(reference.instanceId,reference.labelId,signal);
    return {tenantId:result.tenantId,inventoryId:result.inventoryId,assetId:result.assetId,archived:result.lifecycleState==='archived'};
  }
  async catalog(scope:LabelScope,signal:AbortSignal){
    const [profiles,templates]=await Promise.all([this.client.profiles(scope.tenantId,scope.inventoryId,signal),this.client.templates(scope.tenantId,scope.inventoryId,signal)]);
    if(signal.aborted)throw signal.reason;
    const choices=(profiles??[]).flatMap(profile=>(profile.media??[]).map(media=>{
      const id=`${profile.adapterId}:${media.presetId}:${media.version}`;
      this.media.set(`${scope.tenantId}:${scope.inventoryId}:${id}`,{
        preset_id:media.presetId,version:media.version,width_micrometers:media.widthMicrometers,height_micrometers:media.heightMicrometers,
        margins_micrometers:media.marginsMicrometers,resolution_dpi:media.resolutionDpi,raster_width:media.rasterWidth,raster_height:media.rasterHeight,
        orientation:media.orientation,color_mode:media.colorMode,cut_policy:media.cutPolicy,display_rotation:media.displayRotation,
      });
      return {id,name:media.name};
    }));
    return {profiles:choices,templates:(templates??[]).map(value=>({id:value.id,version:value.version,name:value.name}))};
  }
  async render(scope:LabelScope,choice:LabelChoice,format:'png'|'pdf',signal:AbortSignal){
    const media=this.media.get(`${scope.tenantId}:${scope.inventoryId}:${choice.profileId}`);
    if(!media)throw new Error('media_unavailable');
    await this.client.provision(scope.tenantId,scope.inventoryId,scope.assetId,signal);
    const result=await this.client.render(scope.tenantId,scope.inventoryId,scope.assetId,{media,template:{id:choice.templateId,version:choice.templateVersion,options:{show_reference:choice.showReference}},format},signal);
    const content=await this.client.content(scope.tenantId,scope.inventoryId,result.id,signal);
    return {content,displayRotation:result.displayRotation,width:result.widthPixels,height:result.heightPixels};
  }
}
