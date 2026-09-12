import { useEffect, useRef } from 'react';
import type { SearchBarCommands } from 'react-native-screens';

const SEARCH_DELAY_MS = 300;
/** Native search owns its text; the route owns the applied query. */
export function useExpirationSearch(query: string, onSearch: (query: string) => void) {
 const ref = useRef<SearchBarCommands | null>(null);
 const pending = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
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
 function submit(input: string) {
  text.current = input;
  cancelPending(); const value = input.trim();
  if (value !== applied.current) { applied.current = value; callback.current(value); }
 }
 function change(value: string) {
  text.current = value;
  cancelPending();
  if (!value.trim()) submit('');
  else pending.current = setTimeout(() => submit(value), SEARCH_DELAY_MS);
 }
 function clear() { text.current = ''; ref.current?.clearText(); submit(''); }
 function flush() { submit(text.current); return text.current.trim(); }
 return { ref, change, submit, clear, flush };
}
