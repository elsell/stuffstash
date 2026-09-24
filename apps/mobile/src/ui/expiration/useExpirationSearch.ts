import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useRef, useState } from 'react';

const SEARCH_DELAY_MS = 300;
/** Native search owns its text; the route owns the applied query. */
export function useExpirationSearch(query: string, onSearch: (query: string) => void) {
 const [draft, setDraft] = useState(query);
 const pending = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
 const focused = useRef(false);
 const applied = useRef(query);
 const text = useRef(query);
 const callback = useRef(onSearch);
 callback.current = onSearch;
 function cancelPending() { clearTimeout(pending.current); pending.current = undefined; }
 useEffect(() => {
  if (query !== applied.current) {
   cancelPending(); applied.current = query; text.current = query; setDraft(query);
  }
 }, [query]);
 useEffect(() => () => cancelPending(), []);
 useFocusEffect(useCallback(() => {
  focused.current = true;
  if (text.current.trim() !== applied.current) pending.current = setTimeout(() => submit(text.current), SEARCH_DELAY_MS);
  return () => { focused.current = false; cancelPending(); };
 }, []));
 function submit(input: string) {
  if (!focused.current) return;
  text.current = input; setDraft(input);
  cancelPending(); const value = input.trim();
  if (value !== applied.current) { applied.current = value; callback.current(value); }
 }
 function change(value: string) {
  if (!focused.current) return;
  text.current = value; setDraft(value);
  cancelPending();
  if (!value.trim()) submit('');
  else pending.current = setTimeout(() => submit(value), SEARCH_DELAY_MS);
 }
 function clear() { if (!focused.current) return; text.current = ''; setDraft(''); submit(''); }
 function flush() { submit(text.current); return text.current.trim(); }
 return { draft, change, submit, clear, flush };
}
