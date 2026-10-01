import { t } from '$lib/presentation/localization';
export interface PendingSignInCallbackPresentation {
  title: string;
  description: string;
}

export interface FailedSignInCallbackPresentation extends PendingSignInCallbackPresentation {
  actionLabel: string;
}

export function pendingSignInCallbackPresentation(): PendingSignInCallbackPresentation {
  return {
    title: t('web.signInCallbackPresentation.finishingSecureSignIn'),
    description: t('web.signInCallbackPresentation.stuffStashIsConfirmingYourSession')
  };
}

export function failedSignInCallbackPresentation(_error: unknown): FailedSignInCallbackPresentation {
  return {
    title: t('web.signInCallbackPresentation.weCouldnTFinishSigningYouIn'),
    description: t('web.signInCallbackPresentation.stuffStashCouldnTConfirmYourSessionReturnTo'),
    actionLabel: t('web.signInCallbackPresentation.returnToSignIn')
  };
}
