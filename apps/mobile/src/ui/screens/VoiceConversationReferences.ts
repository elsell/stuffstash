import type { VoiceRealtimeState, VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';

export function voiceConversationReferences(state: VoiceRealtimeState | null): readonly VoiceResponseArtifact[] {
  const references = new Map((state?.responseArtifacts ?? []).map(reference => [reference.assetId, reference]));
  for (const command of state?.actionPlan?.commands ?? []) {
    if (!command.parentAssetId || !command.parentTitle || (command.parentKind !== 'item' && command.parentKind !== 'container' && command.parentKind !== 'location')) continue;
    if (!references.has(command.parentAssetId)) references.set(command.parentAssetId, {
      type: 'asset_reference', assetId: command.parentAssetId, title: command.parentTitle, assetKind: command.parentKind
    });
  }
  return [...references.values()];
}
