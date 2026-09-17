type ReturnNavigation = {
  readonly canGoBack: () => boolean;
  readonly back: () => void;
  readonly replace: (path: '/') => void;
};

export function returnToPreviousOrHome(navigation: ReturnNavigation): void {
  if (navigation.canGoBack()) navigation.back();
  else navigation.replace('/');
}
