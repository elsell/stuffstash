import { Button, Host, HStack, Image, List, Section, Spacer, Text, VStack } from '@expo/ui/swift-ui';
import { accessibilityLabel, accessibilityValue, buttonStyle, contentShape, shapes, disabled, font, foregroundStyle, frame, listStyle } from '@expo/ui/swift-ui/modifiers';
import { spacing } from '../theme/tokens';
import { useFocusedSheetActions } from './useFocusedSheetActions';
import type { MoveSelectionListProps, MoveSelectionRowModel, MoveSelectionStatus } from './MoveSelectionList.types';

const symbols = { root: 'tray', location: 'house', container: 'shippingbox', item: 'cube' } as const;
const secondary = foregroundStyle({ type: 'hierarchical', style: 'secondary' });

const readableSelectionWidth = 720;

/** System List owns scrolling, section spacing, separators and row insets. */
export function MoveSelectionList(props: MoveSelectionListProps) {
  const subject = <VStack alignment="leading" spacing={spacing.md}>
    <VStack alignment="leading" spacing={spacing.xs}>
      <Text modifiers={[secondary]}>{props.subjectLabel}</Text>
      <Text modifiers={[font({ weight: 'semibold' }), foregroundStyle({ type: 'hierarchical', style: 'primary' })]}>{props.subject}</Text>
      <Text modifiers={[secondary]}>{props.context}</Text>
    </VStack>
    <Text>{props.retainedSelection ? 'Selected' : props.title}</Text>
  </VStack>;
  return <Host style={{ flex: 1, width: '100%', maxWidth: readableSelectionWidth, alignSelf: 'center' }}>
    <List modifiers={[listStyle('insetGrouped')]}>
      {props.retainedSelection ? <Section header={subject}><Choice row={props.retainedSelection} /></Section> : null}
      <Section title={props.retainedSelection ? props.title : undefined} header={props.retainedSelection ? undefined : subject}>
        {props.statuses?.map((status, index) => <Status key={index} status={status} />)}
        {props.rows.map(row => <Choice key={row.id} row={row} />)}
      </Section>
    </List>
  </Host>;
}
function Choice({ row }: { readonly row: MoveSelectionRowModel }) {
  const actions = useFocusedSheetActions({ primaryLabel: row.accessibilityLabel, secondaryLabel: '',
    disabled: !!row.disabled, onApply: row.onPress, onBack: () => {} });
  return <Button onPress={actions.onApply} modifiers={[buttonStyle('plain'), disabled(!!row.disabled),
    accessibilityLabel(row.accessibilityLabel), accessibilityValue(row.selected ? 'Selected' : 'Not selected')]}>
    <HStack spacing={spacing.md} modifiers={[contentShape(shapes.rectangle())]}>
      <Image systemName={symbols[row.kind]} modifiers={[secondary, frame({ width: 24 })]} />
      <VStack alignment="leading" spacing={spacing.xs}>
        <Text>{row.label}</Text>
        <Text modifiers={[secondary]}>{row.context}</Text>
      </VStack>
      <Spacer />
      {row.selected ? <Image systemName="checkmark" /> : null}
    </HStack>
  </Button>;
}

function Status({ status }: { readonly status: MoveSelectionStatus }) {
  const actions = useFocusedSheetActions({ primaryLabel: status.retry?.label ?? '', secondaryLabel: '',
    disabled: !status.retry || !!status.retry.disabled, onApply: () => status.retry?.onPress(), onBack: () => {} });
  return <VStack alignment="leading" spacing={spacing.md}>
    <>{status.title ? <Text modifiers={[font({ weight: 'semibold' })]}>{status.title}</Text> : null}<Text modifiers={[secondary]}>{status.message}</Text></>
    {status.retry ? <Button onPress={actions.onApply} modifiers={[buttonStyle('bordered'),
      disabled(actions.disabled), accessibilityLabel(status.retry.label)]}><Text>{status.retry.label}</Text></Button> : null}
  </VStack>;
}
