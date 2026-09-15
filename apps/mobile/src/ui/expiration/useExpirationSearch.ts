import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useRef } from 'react';
import type { SearchBarCommands } from 'react-native-screens';

const SEARCH_DELAY_MS = 300;
/** Native search owns its text; the route owns the applied query. */
export function useExpirationSearch(query: string, onSearch: (query: string) => void) {
 const ref = useRef<SearchBarCommands | null>(null);
 const pending = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
 const focused = useRef(false);
 const applied = useRef(query);
 const text = useRef(query);
 const callback = useRef(onSearch);
 callback.current = onSearch;
 function cancelPending() { clearTimeout(pending.current); pending.current = undefined; }
 useEffect(() => {
  if (query !== applied.current) {
   cancelPending(); applied.current = query; text.current = query; ref.current?.setText(query);
  }
 }, [query]);
 useEffect(() => { ref.current?.setText(text.current); return () => cancelPending(); }, []);
 useFocusEffect(useCallback(() => {
  focused.current = true;
  ref.current?.setText(text.current);
  if (text.current.trim() !== applied.current) pending.current = setTimeout(() => submit(text.current), SEARCH_DELAY_MS);
  return () => { focused.current = false; cancelPending(); };
 }, []));
 function submit(input: string) {
  if (!focused.current) return;
  text.current = input;
  cancelPending(); const value = input.trim();
  if (value !== applied.current) { applied.current = value; callback.current(value); }
 }
 function change(value: string) {
  if (!focused.current) return;
  text.current = value;
  cancelPending();
  if (!value.trim()) submit('');
  else pending.current = setTimeout(() => submit(value), SEARCH_DELAY_MS);
 }
 function clear() { if (!focused.current) return; text.current = ''; ref.current?.clearText(); submit(''); }
 function flush() { submit(text.current); return text.current.trim(); }
 return { ref, change, submit, clear, flush };
}
