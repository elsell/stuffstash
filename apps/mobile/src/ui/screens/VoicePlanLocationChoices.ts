import type { VoiceActionPlanCommand } from '../../application/voice/RealtimeVoiceSession';
import type { VoicePlanCommandDrafts, VoicePlanParentDraft } from './VoicePlanEdits';

const canEdit = (command: VoiceActionPlanCommand) => !!command.id &&
  (command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location');

export function voicePlanLocationChoices(commands: readonly VoiceActionPlanCommand[], commandId: string,
  drafts: VoicePlanCommandDrafts, query: string) {
  const index = commands.findIndex(command => command.id === commandId);
  const command = commands[index];
  const title = (item: VoiceActionPlanCommand) => drafts[item.id!]?.title ?? item.title ?? item.summary;
  const originalParent = commands.find(item => item.id === command?.parentCommandId);
  const current: VoicePlanParentDraft = drafts[commandId]?.parent ?? (command?.parentAssetId
    ? { kind: 'asset', id: command.parentAssetId, label: command.parentTitle ?? 'Existing location' }
    : command?.parentCommandId
      ? { kind: 'command', id: command.parentCommandId, label: originalParent ? title(originalParent) : 'Proposed location' }
      : { kind: 'root', label: 'Inventory root' });
  const search = query.trim().toLocaleLowerCase();
  const proposed = commands.slice(0, Math.max(0, index))
    .filter(item => canEdit(item) && title(item).toLocaleLowerCase().includes(search))
    .map(item => ({ kind: 'command' as const, id: item.id!, label: title(item) }));
  return { valid: !!command && canEdit(command), current, proposed };
}
