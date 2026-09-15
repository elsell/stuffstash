import { useLayoutEffect, useState, type ReactNode } from 'react';
import { subscribeNavigationOptions } from './navigation';

/** Model navigation context updates, failing promptly if options never settle. */
export function NavigationOptionFeedback({ render }: { render: () => ReactNode }) {
  const [, update] = useState(0);
  useLayoutEffect(() => {
    let updates = 0;
    return subscribeNavigationOptions(() => {
      if (++updates > 25) throw new Error('Native header options did not settle');
      update(value => value + 1);
    });
  }, []);
  return render();
}
