import { voiceConversationReferences } from './VoiceConversationReferences';
import { useEffect, useRef, useState } from 'react';
import { AccessibilityInfo, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { VoiceRealtimeState, VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import { AssetCard } from '../components/AssetCard';
import { useAppServices } from '../navigation/AppServicesContext';
import { useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing, radius } from '../theme/tokens';
import { VoiceResponseEntityText } from './VoiceResponseEntityText';

type OpenReference = (reference: VoiceResponseArtifact) => void;
export function VoiceConversationExchange({ exchange, railKey, onOpen }: { readonly exchange: VoiceRealtimeState; readonly railKey: string; readonly onOpen: OpenReference }) {
  const { retryRealtimeActionPlanPhotos } = useVoiceInteractionState();
  const colors = useAppearancePalette();
  return <View style={[styles.exchange, { borderBottomColor: colors.border }]}>
    {exchange.startsNewContext ? <Text style={{ color: colors.textMuted }}>New conversation context</Text> : null}
    {exchange.transcript ? <View style={[styles.user, { backgroundColor: colors.surfaceMuted }]}>
      <Text style={{ color: colors.textMuted }}>You</Text>
      <VoiceResponseEntityText enabled onOpen={onOpen} showFallbackReferences={false} references={voiceConversationReferences(exchange)} text={exchange.transcript} />
    </View> : null}
    {exchange.spokenResponse ? <VoiceResponseEntityText markdown enabled onOpen={onOpen} references={voiceConversationReferences(exchange)} text={exchange.spokenResponse} /> : null}
    {exchange.errorMessage ? <Text selectable style={{ color: colors.warning }}>{exchange.errorMessage}</Text> : null}
    {exchange.actionPlan ? <Text selectable style={{ color: colors.text }}>{exchange.actionPlan.commands.map(command => command.title ?? command.summary).join(', ')} · {exchange.actionPlan.status === 'executed' ? 'Saved' : exchange.actionPlan.status}</Text> : null}
    {exchange.photoAttachmentStatus ? <Text style={{ color: colors.textMuted }}>{exchange.photoAttachmentStatus.message}</Text> : null}
    {exchange.photoAttachmentStatus?.canRetry && exchange.actionPlan ? <Pressable accessibilityRole="button" onPress={() => { void retryRealtimeActionPlanPhotos(exchange.actionPlan!.planId); }} style={styles.control}><Text style={{ color: colors.action }}>Retry photos</Text></Pressable> : null}
    <VoiceResultRail references={voiceConversationReferences(exchange)} railKey={railKey} onOpen={onOpen} />
  </View>;
}
const cardWidth = 264;
export function VoiceResultRail({ references, railKey, onOpen }: { readonly references: readonly VoiceResponseArtifact[]; readonly railKey: string; readonly onOpen: OpenReference }) {
  const bounded = references.filter((reference, index) => references.findIndex(candidate => candidate.assetId === reference.assetId) === index).slice(0, 12);
  const identity = `${railKey}:${bounded.map(reference => reference.assetId).join(',')}`;
  return <VoiceResultRailContent key={identity} references={bounded} identity={identity} onOpen={onOpen} />;
}
function VoiceResultRailContent({ references: bounded, identity, onOpen }: { readonly references: readonly VoiceResponseArtifact[]; readonly identity: string; readonly onOpen: OpenReference }) {
  const { railOffsets } = useVoiceInteractionState();
  const [reduceMotion, setReduceMotion] = useState(true);
  useEffect(() => {
    let active = true;
    void AccessibilityInfo.isReduceMotionEnabled().then(value => { if (active) setReduceMotion(value); }).catch(() => { /* Keep motion disabled when preference is unavailable. */ });
    const subscription = AccessibilityInfo.addEventListener('reduceMotionChanged', setReduceMotion);
    return () => { active = false; subscription.remove(); };
  }, []);
  const initialOffset = useRef({ x: (railOffsets.current[identity] ?? 0) * cardWidth, y: 0 });
  const [viewportWidth, setViewportWidth] = useState(cardWidth);
  const [position, setPosition] = useState(() => railOffsets.current[identity] ?? 0);
  const scroll = useRef<ScrollView>(null);
  const colors = useAppearancePalette();
  if (!bounded.length) return null;
  const current = Math.min(position, bounded.length - 1);
  const move = (next: number) => {
    railOffsets.current[identity] = next; setPosition(next);
    scroll.current?.scrollTo({ x: next * cardWidth, animated: !reduceMotion });
  };
  return <View style={styles.rail}>
    <ScrollView ref={scroll} onLayout={event => setViewportWidth(event.nativeEvent.layout.width)} style={{ flexGrow: 0 }} contentContainerStyle={{ alignItems: 'flex-start', paddingRight: Math.max(0, viewportWidth - cardWidth) }} horizontal snapToInterval={cardWidth} decelerationRate="fast" showsHorizontalScrollIndicator={false}
      contentOffset={initialOffset.current}
      onMomentumScrollEnd={event => { const next = Math.max(0, Math.min(bounded.length - 1, Math.round(event.nativeEvent.contentOffset.x / cardWidth))); railOffsets.current[identity] = next; setPosition(next); }}>
      {bounded.map(reference => <View key={reference.assetId} style={styles.card}><VoiceResultCard reference={reference} onOpen={onOpen} /></View>)}
    </ScrollView>
    {bounded.length > 1 ? <View style={styles.controls}>
      <Pressable accessibilityRole="button" accessibilityLabel="Previous asset card" disabled={current === 0} onPress={() => move(current - 1)} style={styles.control}><Text style={{ color: current === 0 ? colors.textMuted : colors.action }}>Previous</Text></Pressable>
      <Text style={{ color: colors.textMuted }}>{current + 1} of {bounded.length}</Text>
      <Pressable accessibilityRole="button" accessibilityLabel="Next asset card" disabled={current === bounded.length - 1} onPress={() => move(current + 1)} style={styles.control}><Text style={{ color: current === bounded.length - 1 ? colors.textMuted : colors.action }}>Next</Text></Pressable>
    </View> : null}
  </View>;
}
function VoiceResultCard({ reference, onOpen }: { readonly reference: VoiceResponseArtifact; readonly onOpen: OpenReference }) {
  const { assetDetailQuery } = useAppServices();
  const colors = useAppearancePalette();
  const detail = useMobileInventoryServerQuery({ key: (scope, tenant, inventory) => ['voice-card', scope, tenant, inventory, reference.assetId], query: signal => assetDetailQuery.execute(reference.assetId, { signal }) });
  if (!detail.data) return <View style={[styles.placeholder, { backgroundColor: colors.surfaceMuted }]}><Text style={{ color: colors.textMuted }}>{detail.isError ? 'Asset unavailable' : 'Loading asset…'}</Text></View>;
  return <AssetCard asset={detail.data} density="row" style={{ backgroundColor: colors.surfaceMuted, borderRadius: radius.lg, padding: spacing.sm, alignItems: 'flex-start' }} showTags={false} onPress={() => onOpen(reference)} onParentLocationPress={parent => onOpen({ type: 'asset_reference', assetId: parent.id, title: parent.title, assetKind: 'location' })} />;
}
const styles = StyleSheet.create({
  exchange: { gap: spacing.md, paddingBottom: spacing.md, borderBottomWidth: StyleSheet.hairlineWidth },
  user: { alignSelf: 'flex-end', maxWidth: '94%', padding: spacing.sm, borderRadius: radius.lg, gap: spacing.xs },
  rail: { gap: spacing.xs }, card: { width: cardWidth, paddingRight: spacing.sm },
  placeholder: { minHeight: 88, padding: spacing.md, borderRadius: radius.lg },
  controls: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  control: { minHeight: 44, justifyContent: 'center', paddingHorizontal: spacing.sm }
});
