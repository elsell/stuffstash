import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import {
  ProviderProfileSummary,
  VoiceProviderConfiguration,
  VoiceProviderSlot
} from '../../application/providerProfiles/ProviderProfileRepository';
import { colors, radius, spacing } from '../theme/tokens';
import {
  formatProviderProfileCredentialStatusLabel,
  formatProviderProfileLifecycleLabel,
  formatProviderProfileTestStatusLabel,
  formatVoiceProviderCapabilityLabel,
  formatVoiceProviderReadinessLabel,
  formatVoiceProviderSelectionSourceLabel,
  voiceProviderSetupIssueLabels
} from './ProviderProfilesVoiceSetupPresentation';

export function VoiceSetupPanel({
  configuration,
  profiles,
  workingAction,
  onEditCredential,
  onSelectSlotProfile,
  onTestProfile
}: {
  readonly configuration: VoiceProviderConfiguration;
  readonly profiles: readonly ProviderProfileSummary[];
  readonly workingAction?: string;
  readonly onEditCredential: (profile: ProviderProfileSummary) => void;
  readonly onSelectSlotProfile: (
    configuration: VoiceProviderConfiguration,
    slot: VoiceProviderSlot,
    profile: ProviderProfileSummary
  ) => void;
  readonly onTestProfile: (profile: ProviderProfileSummary) => void;
}) {
  return (
    <View style={styles.setupStack}>
      {configuration.slots.map((slot, index) => (
        <VoiceSlotCard
          configuration={configuration}
          index={index}
          key={slot.capability}
          profiles={profiles.filter((profile) => profile.capability === slot.capability)}
          slot={slot}
          workingAction={workingAction}
          onEditCredential={onEditCredential}
          onSelectSlotProfile={onSelectSlotProfile}
          onTestProfile={onTestProfile}
        />
      ))}
    </View>
  );
}

function VoiceSlotCard({
  configuration,
  index,
  profiles,
  slot,
  workingAction,
  onEditCredential,
  onSelectSlotProfile,
  onTestProfile
}: {
  readonly configuration: VoiceProviderConfiguration;
  readonly index: number;
  readonly profiles: readonly ProviderProfileSummary[];
  readonly slot: VoiceProviderSlot;
  readonly workingAction?: string;
  readonly onEditCredential: (profile: ProviderProfileSummary) => void;
  readonly onSelectSlotProfile: (
    configuration: VoiceProviderConfiguration,
    slot: VoiceProviderSlot,
    profile: ProviderProfileSummary
  ) => void;
  readonly onTestProfile: (profile: ProviderProfileSummary) => void;
}) {
  const selectedProfile = slot.selectedProfile;
  const alternatives = profiles.filter((profile) => profile.id !== slot.selectedProfileId);
  const issueLabels = voiceProviderSetupIssueLabels(slot.readiness);

  return (
    <View style={styles.slotCard}>
      <View style={styles.slotHeader}>
        <View style={styles.slotNumber}>
          <Text style={styles.slotNumberText}>{index + 1}</Text>
        </View>
        <View style={styles.slotTitleGroup}>
          <Text style={styles.slotTitle}>{slot.label}</Text>
          <Text style={styles.slotSubtitle}>
            {formatVoiceProviderCapabilityLabel(slot.capability)} / {formatVoiceProviderSelectionSourceLabel(slot.selectionSource)}
          </Text>
        </View>
        <ReadinessPill readiness={slot.readiness} />
      </View>

      {selectedProfile ? (
        <View style={styles.selectedProfileBox}>
          <Text style={styles.selectedProfileName}>{selectedProfile.displayName}</Text>
          <Text style={styles.selectedProfileMeta}>{selectedProfile.providerKind} / {selectedProfile.modelName || 'No model'}</Text>
          <View style={styles.slotFacts}>
            <InfoRow label="Credential" value={formatProviderProfileCredentialStatusLabel(selectedProfile.credentialStatus)} />
            <InfoRow label="Last tested" value={formatProviderProfileTestStatusLabel(selectedProfile.lastTestedAt)} />
          </View>
        </View>
      ) : (
        <View style={styles.emptySlotBox}>
          <Text style={styles.emptyTitle}>No selected profile</Text>
          <Text style={styles.emptyText}>Choose or add a {formatVoiceProviderCapabilityLabel(slot.capability)} profile.</Text>
        </View>
      )}

      {issueLabels.length > 0 ? (
        <View style={styles.issueList}>
          {issueLabels.map((issue) => (
            <Text key={issue} style={styles.issueText}>{issue}</Text>
          ))}
        </View>
      ) : null}

      {selectedProfile && slot.readiness === 'credential_missing' ? (
        <SecondaryActionButton label="Replace credential" onPress={() => onEditCredential(selectedProfile)} />
      ) : null}
      {selectedProfile && slot.readiness === 'untested' ? (
        <SecondaryActionButton label="Test selected profile" onPress={() => onTestProfile(selectedProfile)} />
      ) : null}

      {slot.duplicateProfiles.length > 0 ? (
        <View style={styles.choiceGroup}>
          <Text style={styles.choiceTitle}>Ready duplicates</Text>
          {slot.duplicateProfiles.map((profile) => (
            <ChoiceRow
              isSelected={profile.id === slot.selectedProfileId}
              key={profile.id}
              profile={profile}
              workingAction={workingAction === `select-${slot.capability}-${profile.id}`}
              onSelect={() => onSelectSlotProfile(configuration, slot, profile)}
            />
          ))}
        </View>
      ) : alternatives.length > 0 ? (
        <View style={styles.choiceGroup}>
          <Text style={styles.choiceTitle}>Other {formatVoiceProviderCapabilityLabel(slot.capability)} profiles</Text>
          {alternatives.map((profile) => (
            <ChoiceRow
              isSelected={false}
              key={profile.id}
              profile={profile}
              workingAction={workingAction === `select-${slot.capability}-${profile.id}`}
              onSelect={() => onSelectSlotProfile(configuration, slot, profile)}
            />
          ))}
        </View>
      ) : null}
    </View>
  );
}

function ChoiceRow({
  isSelected,
  profile,
  workingAction,
  onSelect
}: {
  readonly isSelected: boolean;
  readonly profile: ProviderProfileSummary;
  readonly workingAction: boolean;
  readonly onSelect: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      disabled={isSelected || workingAction}
      onPress={onSelect}
      style={styles.choiceRow}
    >
      <View style={styles.choiceTextGroup}>
        <Text style={styles.choiceName}>{profile.displayName}</Text>
        <Text style={styles.choiceMeta}>
          {formatProviderProfileLifecycleLabel(profile.lifecycleState)} / {formatProviderProfileCredentialStatusLabel(profile.credentialStatus)} / {formatProviderProfileTestStatusLabel(profile.lastTestedAt)}
        </Text>
      </View>
      {workingAction ? (
        <ActivityIndicator color={colors.action} />
      ) : (
        <Text style={[styles.choiceAction, isSelected ? styles.choiceActionSelected : null]}>
          {isSelected ? 'Selected' : 'Use'}
        </Text>
      )}
    </Pressable>
  );
}

function SecondaryActionButton({
  label,
  onPress
}: {
  readonly label: string;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={styles.secondaryActionButton}
    >
      <Text style={styles.secondaryActionText}>{label}</Text>
    </Pressable>
  );
}

function ReadinessPill({ readiness }: { readonly readiness: string }) {
  const isReady = readiness === 'ready';
  return (
    <View style={[styles.readinessPill, isReady ? styles.readinessPillReady : null]}>
      <Text style={[styles.readinessPillText, isReady ? styles.readinessPillTextReady : null]}>
        {formatVoiceProviderReadinessLabel(readiness)}
      </Text>
    </View>
  );
}

function InfoRow({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <View style={styles.infoRow}>
      <Text style={styles.infoLabel}>{label}</Text>
      <Text style={styles.infoValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  setupStack: {
    gap: spacing.md
  },
  slotCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  slotHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.md
  },
  slotNumber: {
    alignItems: 'center',
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.sm,
    height: 34,
    justifyContent: 'center',
    width: 34
  },
  slotNumberText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  slotTitleGroup: {
    flex: 1,
    minWidth: 0
  },
  slotTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  slotSubtitle: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 2
  },
  readinessPill: {
    backgroundColor: colors.warningSurface,
    borderColor: colors.brandAmber,
    borderRadius: radius.sm,
    borderWidth: 1,
    maxWidth: 130,
    paddingHorizontal: spacing.sm,
    paddingVertical: 6
  },
  readinessPillReady: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderColor: colors.accent
  },
  readinessPillText: {
    color: colors.warning,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textAlign: 'center'
  },
  readinessPillTextReady: {
    color: colors.text
  },
  selectedProfileBox: {
    backgroundColor: colors.background,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  selectedProfileName: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  selectedProfileMeta: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  slotFacts: {
    marginTop: spacing.sm
  },
  emptySlotBox: {
    backgroundColor: colors.background,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20,
    marginTop: spacing.xs
  },
  issueList: {
    backgroundColor: colors.warningSurface,
    borderColor: colors.brandAmber,
    borderRadius: radius.md,
    borderWidth: 1,
    marginTop: spacing.sm,
    padding: spacing.sm
  },
  issueText: {
    color: colors.warning,
    fontSize: 13,
    fontWeight: '800',
    lineHeight: 18
  },
  choiceGroup: {
    marginTop: spacing.md
  },
  choiceTitle: {
    color: colors.text,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  choiceRow: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.xs,
    minHeight: 58,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  choiceTextGroup: {
    flex: 1,
    minWidth: 0
  },
  choiceName: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  choiceMeta: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 2
  },
  choiceAction: {
    color: colors.action,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  choiceActionSelected: {
    color: colors.textMuted
  },
  secondaryActionButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    justifyContent: 'center',
    marginTop: spacing.sm,
    minHeight: 44,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  secondaryActionText: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  infoRow: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between',
    paddingVertical: spacing.xs
  },
  infoLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0
  },
  infoValue: {
    color: colors.text,
    flex: 1,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textAlign: 'right'
  }
});
