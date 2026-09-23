import { createContext, useContext, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { router } from 'expo-router';

type Presentation = { readonly content: ReactNode; readonly blocked: boolean; readonly onClose: () => void };
type Task = { readonly content: ReactNode; readonly blocked: boolean; readonly cancel: () => void };
type OwnedTask = Task & { readonly owner: object };
type Commands = { show(owner: object, task: Task): void; clear(owner: object): void };
const CommandsContext = createContext<Commands | undefined>(undefined);
const TaskContext = createContext<OwnedTask | undefined>(undefined);
export function AddDestinationTaskProvider({ children }: { readonly children: ReactNode }) {
  const [task, setTask] = useState<OwnedTask>();
  const commands = useMemo<Commands>(() => ({ show: (owner, next) => setTask({ ...next, owner }), clear: owner => setTask(current => current?.owner === owner ? undefined : current) }), []);
  return <CommandsContext.Provider value={commands}><TaskContext.Provider value={task}>{children}</TaskContext.Provider></CommandsContext.Provider>;
}
export function useAddDestinationTask() { return useContext(TaskContext); }
/** Add retains the mutation and draft; this provider only presents one owned visit. */
export function useAddDestinationPresentation(task: Presentation | undefined) {
  const commands = useContext(CommandsContext);
  const owner = useRef({}).current;
  const current = useRef<Presentation | undefined>(undefined);
  const visit = useRef<object | undefined>(undefined);
  useLayoutEffect(() => {
    current.current = task;
    if (!task) { visit.current = undefined; commands?.clear(owner); return; }
    if (!commands) throw new Error('Add destination requires its presentation provider');
    const opening = !visit.current;
    const session = visit.current ?? {};
    visit.current = session;
    commands.show(owner, { content: task.content, blocked: task.blocked, cancel: () => {
      if (visit.current === session && current.current && !current.current.blocked) current.current.onClose();
    } });
    if (opening) router.push('/add-destination');
  }, [commands, owner, task]);
  useLayoutEffect(() => () => { current.current = undefined; visit.current = undefined; commands?.clear(owner); }, [commands, owner]);
}
