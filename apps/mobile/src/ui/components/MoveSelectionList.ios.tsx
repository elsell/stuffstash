import { Button, Host, HStack, Image, List, RNHostView, Section, Spacer, Text, VStack } from '@expo/ui/swift-ui';
import { accessibilityLabel, accessibilityValue, buttonStyle, disabled, font, foregroundStyle, frame, listStyle } from '@expo/ui/swift-ui/modifiers';
import { View } from 'react-native';
import { useFocusedSheetActions } from './useFocusedSheetActions';
import type { MoveSelectionListProps, MoveSelectionRowModel } from './MoveSelectionList.types';

const symbols = { root: 'tray', location: 'house', container: 'shippingbox', item: 'cube' } as const;
const secondary = foregroundStyle({ type: 'hierarchical', style: 'secondary' });

/** System List owns scrolling, section spacing, separators and row insets. */
export function MoveSelectionList(props: MoveSelectionListProps) {
  return <Host style={{ flex: 1 }}>
    <List modifiers={[listStyle('insetGrouped')]}>
      <Section title={props.subjectLabel}>
        <VStack alignment="leading">
          <Text modifiers={[font({ weight: 'semibold' })]}>{props.subject}</Text>
          <Text modifiers={[secondary]}>{props.context}</Text>
        </VStack>
      </Section>
      {props.retainedSelection ? <Section title="Selected"><Choice row={props.retainedSelection} /></Section> : null}
      <Section title={props.title}>
        {props.status ? <RNHostView matchContents><View>{props.status}</View></RNHostView> : null}
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
    <HStack>
      <Image systemName={symbols[row.kind]} modifiers={[secondary, frame({ width: 24 })]} />
      <VStack alignment="leading">
        <Text>{row.label}</Text>
        <Text modifiers={[secondary]}>{row.context}</Text>
      </VStack>
      <Spacer />
      {row.selected ? <Image systemName="checkmark" /> : null}
    </HStack>
  </Button>;
}
