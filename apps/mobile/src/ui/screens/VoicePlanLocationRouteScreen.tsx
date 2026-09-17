import { useMemo, useState } from 'react';
import { router, Stack } from 'expo-router';
import type { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { ScrollView, Text } from 'react-native';
import { useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { useParentCandidates } from '../serverState/useParentCandidates';
import { voicePlanLocationChoices } from './VoicePlanLocationChoices';
import { VoicePlanLocationScreen } from './VoicePlanLocationScreen';
import { useSettingsListStyles } from './SettingsList';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';

export function VoicePlanLocationRouteScreen({ params, parentLookupQuery }: {
  readonly params: { readonly planId: string; readonly commandId: string; readonly scope: string };
  readonly parentLookupQuery: Pick<ParentLookupQuery, 'execute'>;
}) {
  const { state, scopeIdentity, commandDraftState, setCommandDraftState } = useVoiceInteractionState();
  const [query, setQuery] = useState('');
  const { palette, styles } = useSettingsListStyles();
  const plan = state.status === 'ready' ? state.realtime?.actionPlan : undefined;
  const drafts = commandDraftState.planId === params.planId ? commandDraftState.drafts : {};
  const choices = voicePlanLocationChoices(plan?.commands ?? [], params.commandId, drafts, query);
  const valid = choices.valid && plan?.status === 'proposed' && !(state.status === 'ready' && state.realtime?.reviewDecisionPending) && plan.planId === params.planId && scopeIdentity === params.scope;
  const begin = useTaskPresentation(undefined, JSON.stringify([scopeIdentity, params.planId, params.commandId, valid]));
  const candidates = useParentCandidates(query, parentLookupQuery, valid);
  const back = () => { if (!begin()()) return; if (router.canGoBack()) router.back(); else router.replace('/voice'); };
  const backOptions = useNativeHeaderActionOptions([{ kind: 'back', label: 'Back to conversation', onPress: back }], 'left');
  const headerOptions = useMemo(() => ({ title: 'Containing location', headerBackVisible: false, ...backOptions,
    ...(!valid ? { headerSearchBarOptions: undefined } : {}) }), [backOptions, valid]);
  if (!valid) return <>
    <Stack.Screen options={headerOptions} />
    <ScrollView contentInsetAdjustmentBehavior="automatic" style={{ flex: 1, backgroundColor: palette.background }}>
      <Text style={styles.errorMessage}>This proposal is no longer available for editing.</Text>
      <NativeCommandButton label="Back to conversation" onPress={back} />
    </ScrollView>
  </>;
  return <><Stack.Screen options={headerOptions} /><VoicePlanLocationScreen {...choices} matches={candidates.data ?? []} query={query}
    loading={!candidates.isError && candidates.data === undefined} error={candidates.isError}
    onQuery={setQuery} onRetry={() => { if (begin()()) void candidates.refetch(); }}
    onSelect={parent => {
      const isCurrent = begin();
      if (!isCurrent()) return;
      setCommandDraftState(current => {
        if (current.planId && current.planId !== params.planId) return current;
        const existing = current.planId === params.planId ? current.drafts : {};
        return { planId: params.planId, drafts: { ...existing, [params.commandId]: { ...existing[params.commandId], parent } } };
      });
      back();
    }} /></>;
}
