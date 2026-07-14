import type { VoiceActionPlanCommandEdit } from '../../application/voice/RealtimeVoiceSession';

export type VoicePlanParentDraft =
  | { readonly kind: 'root'; readonly label: 'Inventory root' }
  | { readonly kind: 'asset' | 'command'; readonly id: string; readonly label: string };

export type VoicePlanCommandDraft = {
  readonly title?: string;
  readonly parent?: VoicePlanParentDraft;
};

export type VoicePlanCommandDrafts = Readonly<Record<string, VoicePlanCommandDraft>>;

export function voicePlanCommandEdits(drafts: VoicePlanCommandDrafts): readonly VoiceActionPlanCommandEdit[] {
  return Object.entries(drafts).flatMap(([commandId, draft]) => {
    const title = draft.title?.replace(/\s+/g, ' ').trim();
    if (!title && !draft.parent) {
      return [];
    }
    return [{
      commandId,
      ...(title ? { title } : {}),
      ...(draft.parent?.kind === 'root'
        ? { parent: { kind: 'root' as const } }
        : draft.parent
          ? { parent: { kind: draft.parent.kind, id: draft.parent.id } }
          : {})
    }];
  });
}
