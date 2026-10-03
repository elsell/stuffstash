export interface LabelScope {tenantId:string;inventoryId:string;assetId:string}
export interface LabelReference {instanceId:string;labelId:string}
export interface LabelDestination extends LabelScope {archived:boolean}
export interface LabelChoice {profileId:string;templateId:string;templateVersion:number;showReference:boolean}
export interface LabelCatalog {profiles:{id:string;name:string}[];templates:{id:string;version:number;name:string}[]}
export interface LabelArtifact {content:Blob;displayRotation:number;width:number;height:number}
