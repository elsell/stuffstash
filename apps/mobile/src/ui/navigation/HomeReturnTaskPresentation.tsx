import { createContext, useContext, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { router } from 'expo-router';

export type HomeReturnTask = {
  readonly content: ReactNode;
  readonly requestClose: () => void;
  readonly focusChanged: (focused: boolean) => void;
};
type OwnedTask = HomeReturnTask & { readonly owner: object };
type Presentation = { readonly show: (owner: object, task: HomeReturnTask) => void; readonly clear: (owner: object) => void };
const Commands = createContext<Presentation | undefined>(undefined);
const CurrentTask = createContext<OwnedTask | undefined>(undefined);

/** UI composition only: the inventory-scoped Home hook retains command ownership. */
export function HomeReturnTaskProvider({ children }: { readonly children: ReactNode }) {
  const [task, setTask] = useState<OwnedTask>();
  const commands = useMemo<Presentation>(() => ({
    show: (owner, next) => setTask({ ...next, owner }),
    clear: owner => setTask(current => current?.owner === owner ? undefined : current)
  }), []);
  return <Commands.Provider value={commands}><CurrentTask.Provider value={task}>{children}</CurrentTask.Provider></Commands.Provider>;
}

export function useHomeReturnTaskPresentation(task: HomeReturnTask | undefined) {
  const commands = useContext(Commands);
  if (!commands) throw new Error('Home return task requires its presentation provider');
  const owner = useRef({}).current;
  const presented = useRef(false);
  useLayoutEffect(() => {
    if (task) {
      commands.show(owner, task);
      if (!presented.current) { presented.current = true; router.push('/home-return-details'); }
    } else {
      commands.clear(owner);
      presented.current = false;
    }
  }, [commands, owner, task]);
  useLayoutEffect(() => () => commands.clear(owner), [commands, owner]);
}

export function useHomeReturnTask() { return useContext(CurrentTask); }
