type URLListener = (event: { url: string }) => void;
const listeners = new Set<URLListener>();

/** Stateful native URL source; subscribers receive the same events as Expo Linking. */
export const labelLinkSourceFake = {
  initialURL: null as string | null,
  receive(url: string) { for (const listener of listeners) listener({ url }); },
  reset() { this.initialURL = null; listeners.clear(); }
};

export async function getInitialURL(): Promise<string | null> { return labelLinkSourceFake.initialURL; }
export function addEventListener(_event: 'url', listener: URLListener) {
  listeners.add(listener);
  return { remove: () => { listeners.delete(listener); } };
}
