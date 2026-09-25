import type { CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';
import { createContext, useCallback, useContext, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { router, useFocusEffect } from 'expo-router';
import { AssetTagSelectionScreen, type AssetTagSelectionOption } from '../screens/AssetTagSelectionScreen';

type Task = { readonly content: ReactNode; readonly cancel: () => void };
type OwnedTask = Task & { readonly owner: object };
type Commands = { show(owner: object, task: Task): void; clear(owner: object): void };
const PresentationCommands = createContext<Commands | undefined>(undefined);
const CurrentTask = createContext<OwnedTask | undefined>(undefined);

export function AssetTagSelectionTaskProvider({ children }: { readonly children: ReactNode }) {
  const [task, setTask] = useState<OwnedTask>();
  const commands = useMemo<Commands>(() => ({
    show: (owner, next) => setTask({ ...next, owner }),
    clear: owner => setTask(current => current?.owner === owner ? undefined : current)
  }), []);
  return <PresentationCommands.Provider value={commands}><CurrentTask.Provider value={task}>{children}</CurrentTask.Provider></PresentationCommands.Provider>;
}
export function useAssetTagSelectionTask() { return useContext(CurrentTask); }

type SelectionOwner = {
  readonly scope: string;
  readonly disabled: boolean;
  readonly tags: readonly AssetTagSelectionOption[];
  readonly selectedIds: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
  readonly onChange: (ids: readonly string[], newTags?: readonly CreateAssetTagDraft[]) => void;
};
/** Navigation carries the presentation only; committed form callbacks own the draft. */
export function useAssetTagSelectionVisit(options: SelectionOwner) {
  const commands = useContext(PresentationCommands);
  const owner = useRef({}).current;
  const focused = useRef(false);
  useFocusEffect(useCallback(() => {
    focused.current = true;
    return () => { focused.current = false; };
  }, []));
  const current = useRef<SelectionOwner | undefined>(undefined);
  const [visit, setVisit] = useState<{ scope: string; ids: readonly string[] }>();
  const active = useRef<typeof visit>(undefined);
  useLayoutEffect(() => {
    current.current = options;
    return () => { current.current = undefined; };
  }, [options]);
  const finish = useCallback((expected: NonNullable<typeof visit>, ids?: readonly string[], newTags?: readonly CreateAssetTagDraft[]) => {
    const latest = current.current;
    const pending = active.current;
    if (!pending || pending !== expected) return;
    active.current = undefined;
    setVisit(undefined);
    if (ids && latest && !latest.disabled && latest.scope === pending.scope) latest.onChange(ids, newTags);
  }, []);
  useLayoutEffect(() => {
    if (!visit || visit.scope !== options.scope || options.disabled) {
      active.current = undefined;
      commands?.clear(owner);
      if (visit) setVisit(undefined);
      return;
    }
    active.current = visit;
    commands?.show(owner, { cancel: () => finish(visit), content: <AssetTagSelectionScreen
      tags={options.tags} initialSelectedIds={visit.ids} initialNewTags={options.newTags} onDone={(ids, newTags) => finish(visit, ids, newTags)} onCancel={() => finish(visit)} /> });
  }, [commands, owner, visit, options.scope, options.disabled, options.tags, options.newTags, finish]);
  useLayoutEffect(() => () => { active.current = undefined; commands?.clear(owner); }, [commands, owner]);
  return useCallback(() => {
    const latest = current.current;
    if (!focused.current || !latest || latest.disabled || active.current) return;
    if (!commands) throw new Error('Tag selection requires its presentation provider');
    const next = { scope: latest.scope, ids: [...latest.selectedIds] };
    active.current = next;
    setVisit(next);
    router.push('/asset-tag-selection');
  }, [commands]);
}
