import { useEffect, useRef } from 'react';
import { usePathname, useRouter } from 'expo-router';
import { usePendingLabel } from './LabelLinkContext';
import { useAppFeedback } from '../feedback/AppFeedback';
import { t } from '../../presentation/localization';
/** Only the exact Home and Browse roots are safe for direct incoming navigation. */
export function PendingLabelNavigation() {
  const pending = usePendingLabel(); const feedback = useAppFeedback(); const router = useRouter(); const path = usePathname();
  const offeredRevision = useRef<number | undefined>(undefined);
  useEffect(() => {
    if ((!pending.reference && !pending.invalid) || offeredRevision.current === pending.revision) return;
    offeredRevision.current = pending.revision;
    if (path === '/scan-label') return;
    if (pending.reference && (path === '/' || path === '/search')) {
      router.push('/scan-label');
      return;
    }
    feedback.showNotice({ tone: 'info', title: t('labels.mobile.scan'), message: t(pending.invalid ? 'labels.mobile.unsupported' : 'labels.mobile.ready'),
      action: { label: t('labels.mobile.open'), onPress: () => router.push('/scan-label') } });
  }, [pending.revision, path, feedback, router]);
  return null;
}
