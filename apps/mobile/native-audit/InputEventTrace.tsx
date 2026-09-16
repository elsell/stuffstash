import { useLayoutEffect, useRef, useState } from 'react';
import { Button, Text, View, type TextInputProps } from 'react-native';

const maximumEvents = 256;
type Entry = { readonly kind: 'key'; readonly key: string }
  | { readonly kind: 'change'; readonly text: string; readonly count: number }
  | { readonly kind: 'selection'; readonly start: number; readonly end: number }
  | { readonly kind: 'commit'; readonly text: string };

// Synthetic runner fixtures only. Recording must not add renders to the typing path.
export function useInputEventTrace(value: string) {
  const events = useRef<Entry[]>([]);
  const dropped = useRef(0);
  const [snapshot, setSnapshot] = useState('');
  function record(entry: Entry) {
    if (events.current.length === maximumEvents) {
      events.current.shift();
      dropped.current += 1;
    }
    events.current.push(entry);
  }
  useLayoutEffect(() => { record({ kind: 'commit', text: value }); }, [value]);
  const onKeyPress: TextInputProps['onKeyPress'] = event => {
    record({ kind: 'key', key: event.nativeEvent.key });
  };
  const onChange: TextInputProps['onChange'] = event => {
    record({ kind: 'change', text: event.nativeEvent.text, count: event.nativeEvent.eventCount });
  };
  const onSelectionChange: TextInputProps['onSelectionChange'] = event => {
    record({ kind: 'selection', ...event.nativeEvent.selection });
  };
  return {
    onKeyPress, onChange, onSelectionChange,
    controls: <View>
      <Button title="Capture input events" onPress={() => setSnapshot(JSON.stringify({ value, dropped: dropped.current, events: events.current }))} />
      {snapshot ? <Text testID="audit-input-event-trace">{snapshot}</Text> : null}
    </View>
  };
}
