import { useAppFeedback } from '../feedback/AppFeedback';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { usePullRefresh } from './usePullRefresh';

/** Explicit refresh feedback belongs to the resource and visit that requested it. */
export function usePullRefreshFeedback({ refresh, resourceKey, failureTitle }: {
  readonly refresh: () => Promise<unknown>;
  readonly resourceKey: readonly unknown[];
  readonly failureTitle: string;
}) {
  const feedback = useAppFeedback();
  const captureVisit = useTaskPresentation(undefined, JSON.stringify(resourceKey));
  return usePullRefresh(async () => {
    const isCurrent = captureVisit();
    try { await refresh(); }
    catch {
      if (isCurrent()) feedback.showNotice({
        tone: 'error', title: failureTitle,
        message: 'Please try refreshing again.'
      });
    }
  });
}
