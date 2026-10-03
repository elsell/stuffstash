import type {AddAssetSubmission} from '$lib/domain/inventory';
import type {CreateAssetWorkflowCheckpoint} from '$lib/application/workspaceAssetWorkflow';
export interface AssetPrintCreationAttempt {draft:AddAssetSubmission;checkpoint:CreateAssetWorkflowCheckpoint}
export function prepareAssetPrintCreation(draft:AddAssetSubmission,key:string):AssetPrintCreationAttempt {
 const {photos,...fields}=draft;
 return {draft:{...JSON.parse(JSON.stringify(fields)),photos:photos.map(photo=>({...photo})),creationKey:key},checkpoint:{dispatchStarted:false,ambiguous:false}};
}
